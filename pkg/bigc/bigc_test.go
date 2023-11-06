package bigc_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

// func TestGetLineBySku__PF(t *testing.T) {
// 	pfSource := "/Users/home/Library/CloudStorage/Dropbox-DevelinManagement/Zihan/files/out/231016__web__pf.tsv"
// 	pf, err := parsers.ParseProductFile(pfSource)
// 	if err != nil {log.Fatalln(err)}
// 	pfl, err := pf.GetLineBySku("18787777046")
// 	if err != nil {
// 		fmt.Println("!!!!!", err)
// 	}
// 	fmt.Println(pfl)
// }

func TestGetCats(t *testing.T) {
	c := bigc.MustGetClient()
	retv := make(map[int]string, 0)

	cats, err := c.GetCategories(map[string]string{"parent_id": "0"})
	if err != nil {
		t.Fatal(err)
	}

	for _, cat := range cats {
		retv[cat.ID] = cat.Name
		retv = recurse(c, cat.ID, cat.Name+"/", retv)
	}

	fmt.Println("len retv: ", len(retv))
	fmt.Println(retv)
}

func recurse(c *bigc.BigCommerceClient, id int, path string, m map[int]string) map[int]string {
	kids, err := c.GetCategories(map[string]string{"parent_id": fmt.Sprint(id)})
	if err != nil {
		log.Fatal(err)
	}

	if len(kids) == 0 {
		return m
	}

	for _, cat := range kids {
		m[cat.ID] = path + cat.Name + "/"
		m = recurse(c, cat.ID, path+cat.Name+"/", m)
	}

	return m
}

func TestGetAllChildren2(t *testing.T) {
	c := bigc.MustGetClient()
	cat, err := c.GetCategoryFromID(0)
	if err != nil {
		t.Fatal(err)
	}
	cats, err := cat.GetAllChildren2(c)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(cats))
}
