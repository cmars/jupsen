package main

import (
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

var (
	url   = flag.String("url", "localhost:27017", "url of the mongoserver")
	count = flag.Int("count", 6000, "number of list items")
)

type List struct {
	Id      string `bson:"_id"`
	Numbers []int  `bson:"numbers"`
}

type Stats struct {
	Expected []int
	Pass     int
	Fail     int
}

func WriteList(c *mgo.Collection, id string, count int) Stats {
	stats := Stats{}

	for i := 0; i < count; i++ {
		//TODO NO
		err := c.Update(bson.M{"_id": id}, bson.M{"$push": bson.M{"numbers": i}})
		if err != nil {
			stats.Fail += 1
			log.Printf("failed update: %v\n", err)
			newSession := c.Database.Session.Copy()
			c.Database.Session.Close()
			c.Database.Session = newSession
			continue
		}
		stats.Pass += 1
		stats.Expected = append(stats.Expected, i)
		if i%100 == 0 {
			fmt.Println(i)
		}
	}
	return stats
}

func Compare(c *mgo.Collection, id string, stats Stats) {
	//TODO New session here couldn't hurt
	var result List
	err := c.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		log.Fatalf("failed to read results: %v\n", err)
	}
	fmt.Printf("Pass:%d\n", stats.Pass)
	fmt.Printf("Fail:%d\n", stats.Fail)
	fmt.Printf("Expect Len: %d\n", len(stats.Expected))
	fmt.Printf("Actual Len: %d\n", len(result.Numbers))
}

func main() {
	flag.Parse()
	fmt.Printf("Connecting to %s and appending %d times\n", *url, *count)
	session, err := mgo.Dial(*url)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer session.Close()

	session.EnsureSafe(&mgo.Safe{W: 1, FSync: false})
	list := List{
		Id:      "my-list",
		Numbers: []int{},
	}

	c := session.DB("jupsen").C("list")
	err = c.Insert(&list)
	if err != nil {
		log.Fatalf("failed to set starting point: %v\n", err)
	}

	stats := WriteList(c, list.Id, *count)
	Compare(c, list.Id, stats)

	err = c.DropCollection()
	if err != nil {
		log.Fatalf("failed to drop collection: %v", err)
	}
}