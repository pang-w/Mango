package segment


import (
    "fmt"
    "os"
    //"io/ioutil"
    "strings"
    "Mango/management/models"
    "Mango/management/utils"
    //"github.com/qiniu/log"
    "github.com/jason-zou/sego"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

type GuokuSegmenter struct {
    seg sego.Segmenter
}


func (seg *GuokuSegmenter) LoadDictionary() bool {
    //fmt.Println(os.Chdir("/home/jasonz/gocode/src/Mango/management"))
    seg.seg.LoadDictionary("dictionary/dictionary.txt")
    sess, err := mgo.Dial("10.0.1.23")
    if err != nil {
        return false
    }
    dc := sess.DB("words").C("dict_chi_eng")
    words := make([]models.DictWord, 0)
    dc.Find(bson.M{"$or" : []bson.M{bson.M{"freq" : bson.M{"$gt" : 50}, "deleted" : false}, bson.M{"type" : "manual", "deleted" : false}}}).All(&words)
    length := len(words)
    wordUnits := make([]sego.WordUnit, length)
    for i := 0; i < length; i++ {
        wordUnits[i].Word = words[i].Word
        freq := words[i].Freq
        if freq < 200 && words[i].Type == "manual" {
            freq = 200000
        }
        wordUnits[i].Freq = freq
        wordUnits[i].Pos = "n"
    }

    seg.seg.LoadDictionaryFromArray(wordUnits)
    bc := sess.DB("words").C("brands")
    brands := make([]models.BrandsWord, 0)
    bc.Find(bson.M{"$or" : []bson.M{bson.M{"valid" : true}, bson.M{"deleted" : false, "freq" : bson.M{"$gt": 50}}}}).All(&brands)
    length = len(brands)
    brandWords := make([]sego.WordUnit, 0)
    for i := 0; i < length; i++ {
        bs := strings.Split(brands[i].Name, "/")
        for _, v := range bs {
            if len(v) == 0 {
                continue
            }
            freq := brands[i].Freq
            if freq < 100 && brands[i].Valid {
                freq = 100
            }
            brandWords = append(brandWords, sego.WordUnit{Word : v, Freq : freq, Pos : "n"})
        }
    }
    seg.seg.LoadDictionaryFromArray(brandWords)
    seg.seg.CalDistance()
    return true
}


func (seg *GuokuSegmenter) Segment(str string) [][]string {
    nstr := utils.StripPuncsAndSymbols(str)
    sgs := seg.seg.Segment([]byte(nstr))
    return sego.SegmentsToTextSlice(sgs)
}



