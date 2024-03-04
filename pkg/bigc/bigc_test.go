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

func recurse(c *bigc.Client, id int, path string, m map[int]string) map[int]string {
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

func TestGetProductsWithVariantsAndInventoryLevel0(t *testing.T) {
	c := bigc.MustGetClient()
	ps, err := c.GetAllProducts(map[string]string{"include": "variants,images", "inventory_level": "0"})
	if err != nil {
		log.Fatalln(err)
	}

	for _, p := range ps {
		if len(p.Variants) > 0 {
			for _, v := range p.Variants {
				if v.ImageURL == "" {
					continue
				}
				fmt.Println(v.Sku, v.InventoryLevel)
			}
		}
	}

}

func TestGetCustomerById(t *testing.T) {
	c := bigc.MustGetClient()
	cuss, err := c.GetCustomers(map[string]string{"id:in": "3408"})
	if err != nil {
		t.Fatal(err)
	}

	if len(cuss) != 1 {
		t.Fatal("len(cuss) != 1", len(cuss))
	}

	fmt.Printf("%+v\n", cuss[0].CustomerGroupID)

}

func TestUpdateCustomer(t *testing.T) {
	c := bigc.MustGetClient()

	cuss, err := c.GetCustomers(map[string]string{"id:in": "7452"})
	if err != nil {
		t.Fatal(err)
	}

	if len(cuss) != 1 {
		t.Fatal("len(cuss) != 1", len(cuss))
	}

	cus, err := c.UpdateCustomer(cuss[0], 7452, bigc.WithCustomerGroupID(6))
	if err != nil {
		t.Fatal(err)
	}

	if cus.CustomerGroupID != 6 {
		t.Fatal("cus.CustomerGroupID != 6", cus.CustomerGroupID)
	}
}

func TestGetProductsById(t *testing.T) {
	c := bigc.MustGetClient()
	ids := []int{
		10006,
		10020,
		10023,
		10035,
		10045,
		10048,
		10053,
		10054,
		10055,
		10080,
		10081,
		10083,
		10084,
	}

	ps, errs := c.GetProductsById(ids)
	if len(ps) != len(ids) {
		t.Fatal("len(pch) != len(ids)", len(ps), len(ids))
	}
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, p := range ps {
		t.Log(p.Name)
	}

}
