package bigc

import (
	"fmt"
	"slices"

	"github.com/samber/lo"
)

const (
	PROMOTIONS                = 142
	PRODUCTSALE               = 668
	PRODUCTSALE_10            = 1005
	SALE_10                   = 1030
	PRODUCTSALE_20            = 146
	SALE_20                   = 143
	PRODUCTSALE_30            = 732
	SALE_30                   = 838
	PRODUCTSALE_40            = 866
	SALE_40                   = 865
	PRODUCTSALE_50            = 870
	SALE_50                   = 869
	PRODUCTSALE_60            = 1012
	PRODUCTSALE_AUSMADE       = 352
	PRODUCTSALE_VEGAN         = 361
	PRODUCTSALE_NATURAL       = 363
	PRODUCTSALE_REEF          = 665
	PRODUCTSALE_SALE          = 668
	PRODUCTSALE_LIMITED_ED    = 671
	PRODUCTSALE_NEW           = 683
	PRODUCTSALE_SPECIAL       = 725
	PRODUCTSALE_GWP           = 744
	PRODUCTSALE_5_DOLLAR_OFF  = 873
	PRODUCTSALE_WEB_EXCLUSIVE = 888
	PRODUCTSALE_NO_CNC        = 1003
	PRODUCTSALE_PRAC_ONLY     = 1183
	CHRISTMAS_CLEARANCE       = 899
	WEB_EXCLUSIVE             = 887
	PROMO_SET_SALES           = 1051
	PROMO_AUSMADE             = 351
	PROMO_SALE                = 142
	RETIRED_PRODUCTS          = 1230
	NEW                       = 1041
	CLEARANCE                 = 691
)

func getSaleCategoryIDs() []int {
	return []int{
		PROMOTIONS,
		PRODUCTSALE,
		PRODUCTSALE_10,
		SALE_10,
		PRODUCTSALE_20,
		SALE_20,
		PRODUCTSALE_30,
		SALE_30,
		PRODUCTSALE_40,
		SALE_40,
		PRODUCTSALE_50,
		SALE_50,
		PRODUCTSALE_60,
		PRODUCTSALE_AUSMADE,
		PRODUCTSALE_VEGAN,
		PRODUCTSALE_NATURAL,
		PRODUCTSALE_REEF,
		PRODUCTSALE_SALE,
		PRODUCTSALE_LIMITED_ED,
		PRODUCTSALE_NEW,
		PRODUCTSALE_SPECIAL,
		PRODUCTSALE_GWP,
		PRODUCTSALE_5_DOLLAR_OFF,
		PRODUCTSALE_WEB_EXCLUSIVE,
		PRODUCTSALE_NO_CNC,
		PRODUCTSALE_PRAC_ONLY,
		CHRISTMAS_CLEARANCE,
		WEB_EXCLUSIVE,
		PROMO_SET_SALES,
		PROMO_AUSMADE,
		PROMO_SALE,
		CLEARANCE,
	}
}

func RemoveSaleCategories(ids []int) []int {
	saleIds := getSaleCategoryIDs()

	return lo.Filter(ids, func(id int, _ int) bool {
		return !slices.Contains(saleIds, id)
	})
}

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (cat *Category) GetParent(c *BigCommerceClient) (Category, error) {
	parentId := cat.ParentID
	if parentId == 0 {
		return Category{}, nil
	}
	return c.GetCategoryFromID(parentId)
}

func (cat *Category) GetSiblings(c *BigCommerceClient) ([]Category, error) {
	return c.GetCategories(map[string]string{"parent_id": fmt.Sprint(cat.ParentID)})
}

func (cat *Category) GetChildren(c *BigCommerceClient) ([]Category, error) {
	return c.GetCategories(map[string]string{"parent_id": fmt.Sprint(cat.ID)})
}

func (cat *Category) GetAllChildren(c *BigCommerceClient) ([]*Category, error) {
	children, err := cat.GetChildren(c)
	if err != nil {
		return nil, err
	}

	var allChildren []*Category

	for _, child := range children {
		allChildren = append(allChildren, &child)
		grandChildren, err := child.GetAllChildren(c) // recursive call
		if err != nil {
			return nil, err
		}
		allChildren = append(allChildren, grandChildren...)
	}

	return allChildren, nil
}

func (cat *Category) GetAllChildren2(c *BigCommerceClient) ([]Category, error) {
	children, err := cat.GetChildren(c)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		grandChildren, err := child.GetAllChildren2(c) // recursive call
		if err != nil {
			return nil, err
		}
		children = append(children, grandChildren...)
	}

	return children, nil
}

type CategoriesSlice []Category

func (cats CategoriesSlice) Export(path string) error {
	headers, e := GetStructFields(cats[0])
	if e != nil {
		return e
	}

	var content [][]string
	var tmp []string
	for _, p := range cats {
		values, e := GetStructValues(p)
		if e != nil {
			return e
		}
		for _, v := range values {
			tmp = append(tmp, fmt.Sprint(v))
		}
		content = append(content, tmp)
		tmp = []string{}
	}

	return WriteToTsv(path, headers, content)
}

// type TreeNode struct {
// 	id int
// 	name string
// 	children []*TreeNode
// }
//
// type Tree struct {
// 	Root *TreeNode
// }
//
// // NewTree creates a new Tree with the given root node.
// func NewTree(root *TreeNode) *Tree {
// 	return &Tree{
// 		Root: root,
// 	}
// }
//
// // Traverse implements a Depth-First traversal of the tree.
// // It applies a given function to each node it visits.
// func (n *TreeNode) Traverse(f func(node *TreeNode)) {
// 	if n == nil {
// 		return
// 	}
//
// 	f(n) // Apply the function on the current node
// 	for _, child := range n.Children {
// 		child.Traverse(f) // Recursive call on each child
// 	}
// }
