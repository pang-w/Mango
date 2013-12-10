package main

import (
	"Mango/management/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"strconv"
	"sync"
	"time"
)

const THREADSNUM int = 15
const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
)

func main() {

	log.Println("hello")
	shopitem := new(utils.ShopItem)
	minerals := utils.MongoInit(MGOHOST, MGODB, "minerals")
	for {
		err := minerals.Find(bson.M{"state": "posted"}).One(&shopitem)
		if err != nil {
			log.Println(err.Error())
			panic(err)
		}
		shopid := strconv.Itoa(shopitem.Shop_id)
		items := shopitem.Items_list
		run(shopid, items)
	}
}

var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
var mgominer *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "minerals")

func run(shopid string, items []string) {
	var allowchan chan bool = make(chan bool, THREADSNUM)

	log.Printf("\n\nStart to run fetch")
	shoptype := TAOBAO
	if len(items) == 0 {
		log.Println("%s has no items", shopid)
		return
	} else {
		istmall, err := utils.IsTmall(items[0])
		if err != nil {
			log.Println("there is an error during judge")
			log.Println(err.Error())
			return
		}
		if istmall {
			shoptype = TMALL
		}

		var wg sync.WaitGroup
		for _, itemid := range items {
			allowchan <- true
			wg.Add(1)
			go func(itemid string) {
				defer wg.Done()

				defer func() { <-allowchan }()
				log.Printf("start to fetch %s", itemid)
				page, err, detail := utils.Fetch(itemid, shoptype)
				if err != nil {
					log.Printf("%s failed", itemid)

					failed := utils.FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
					err = mgofailed.Insert(&failed)
					if err != nil {
						log.Println(err.Error())
						mgofailed.Update(bson.M{"itemid": itemid}, bson.M{"$set": failed})
					}
				} else {

					log.Printf("%s 成功", itemid)
					info, missing, err := utils.Parse(page, detail, itemid, shopid, shoptype)
					log.Println("解析完毕")
					instock := true
					parsed := false
					if err != nil {
						log.Println(err.Error())
						if missing {
							parsed = true
							instock = false
						} else {
							parsed = false
							if err.Error() == "cattag" {
								//有可能该商品找不到了
								instock = false
							} else {
								failed := utils.FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
								err = mgofailed.Insert(&failed)
								if err != nil {
									log.Println(err.Error())
									mgofailed.Update(bson.M{"itemid": itemid}, bson.M{"$set": failed})
								}
								return
							}
						}
					} else {
						instock = info.InStock
						log.Println("开始发送")
						err = utils.Post(info)
						if err != nil {
							log.Println("发送出现错误")
							log.Println(err.Error())
							parsed = false

						} else {
							log.Println("发送完毕")
							parsed = true
						}
					}
					successpage := utils.Pages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, FontPage: page, UpdateTime: time.Now().Unix(), DetailPage: detail, Parsed: parsed, InStock: instock}
					err = mgopages.Insert(&successpage)
					if err != nil {
						log.Println(err.Error())
						mgopages.Update(bson.M{"itemid": itemid}, bson.M{"$set": successpage})
					}
				}
			}(itemid)
		}
		wg.Wait()
	}
	close(allowchan)
	sid, _ := strconv.Atoi(shopid)
	err := mgominer.Update(bson.M{"shop_id": sid}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
	if err != nil {
		log.Println("update minerals state error")
		log.Println(err.Error())
	}
}

/*
func hello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println(r.Form)
	log.Println("path", r.URL.Path)
	log.Println("scheme", r.URL.Scheme)
	log.Println(r.Form["url_long"])
	for k, v := range r.Form {
		log.Println("key:", k)
		log.Println("val:", strings.Join(v, " "))
	}
	fmt.Fprintf(w, "hello world")
}

func main() {
	go func() {
		http.HandleFunc("/", hello)
		err := http.ListenAndServe(":9090", nil)
		if err != nil {
			log.Fatal("listen and server :", err)
		}
	}()
	fmt.Println("hello")
	select {}
}
*/