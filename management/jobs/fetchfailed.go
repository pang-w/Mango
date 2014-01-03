package main

import (
	"Mango/management/crawler"
	"Mango/management/utils"
	"github.com/qiniu/log"
	"labix.org/v2/mgo/bson"
)

const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
	MANGO   string = "mango"
)

func main() {
	for {
		run()
	}
}

func run() {
	log.SetOutputLevel(log.Ldebug)
	mgopages := utils.MongoInit(MGOHOST, MGODB, "pages")
	mgofailed := utils.MongoInit(MGOHOST, MGODB, "failed")
	mgoMango := utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")
	log.Info("start to refetch")
	iter := mgofailed.Find(nil).Iter()
	failed := new(crawler.FailedPages)
	for iter.Next(&failed) {
		info, err := mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
		if err != nil {
			log.Info(info.Removed)
			log.Info(err.Error())
		}
		page, detail, instock, err := crawler.FetchItem(failed.ItemId, failed.ShopType)
		if err != nil {
			if instock {
				crawler.SaveFailed(failed.ItemId, failed.ShopId, failed.ShopType, mgofailed)
			} else {
				mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
			}
			/*
				if err.Error() == "404" {
					log.Info("start to remove item")
					_, err = mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
					if err != nil {
						log.Info(err.Error())
					}
				} else {
					log.Infof("%s refetch failed\n", failed.ItemId)
					newfail := utils.FailedPages{ItemId: failed.ItemId, ShopId: failed.ShopId, ShopType: failed.ShopType, UpdateTime: time.Now().Unix(), InStock: failed.InStock}
					err = mgofailed.Insert(&newfail)
					if err != nil {
						log.Info(err.Error())
						mgofailed.Update(bson.M{"itemid": failed.ItemId}, bson.M{"$set": newfail})
					}
				}
			*/
		} else {
			log.Info("%s refetch successed", failed.ItemId)
			info, instock, err := crawler.ParsePage(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
			/*
				instock := true
				parsed := false
				if err != nil {
					if missing {
						parsed = true
						instock = false
					} else if err.Error() != "聚划算" {
						parsed = false
						if err.Error() == "cattag" {
							instock = false
						}

					}
					log.Info(err.Error())
				} else {
					instock = info.InStock
					err = utils.Save(info, mgoMango)
					if err != nil {
						log.Info(err.Error())

					}
				}
			*/

			if err != nil {
				continue
			}
			instock = info.InStock
			err = crawler.Save(info, mgoMango)
			if err != nil {
				log.Error(err)
				continue
			}
			crawler.SaveSuccessed(failed.ItemId, failed.ShopId, failed.ShopType, page, detail, true, instock, mgopages)
			//	successpage := crawler.Pages{ItemId: failed.ItemId, ShopId: failed.ShopId, ShopType: failed.ShopType, FontPage: page, DetailPage: detail, UpdateTime: time.Now().Unix(), Parsed: true, InStock: instock}
			//	err = mgopages.Insert(&successpage)
			//	if err != nil {
			//		log.Info(err.Error())
			//		mgopages.Update(bson.M{"itemid": failed.ItemId}, bson.M{"$set": successpage})
			//	}
		}
	}
	if err := iter.Close(); err != nil {
		log.Info(err.Error())
	}
}
