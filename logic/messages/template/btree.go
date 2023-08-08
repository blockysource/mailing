package emailtemplate

import (
	"strings"
	"sync"

	"github.com/google/btree"
)

func newTemplateBTree() *templateBTree {
	return &templateBTree{
		templates: btree.NewG[parsedTemplateBTreeItem](btree.DefaultFreeListSize, func(a, b parsedTemplateBTreeItem) bool {
			return strings.Compare(a.uid, b.uid) < 0
		}),
	}
}

type (
	templateBTree struct {
		sync.RWMutex
		templates *btree.BTreeG[parsedTemplateBTreeItem]
	}
	parsedTemplateBTreeItem struct {
		uid string
		t   *TemplateParser
	}
)

// Get returns the item with the given uid from the tree.
func (t *templateBTree) Get(uid string) (*TemplateParser, bool) {
	t.RLock()
	defer t.RUnlock()

	item, ok := t.templates.Get(parsedTemplateBTreeItem{uid: uid})
	if !ok {
		return nil, false
	}
	return item.t, true
}

// Len returns the number of items in the tree.
func (t *templateBTree) Len() int {
	t.RLock()
	defer t.RUnlock()

	return t.templates.Len()
}

// ReplaceOrInsert adds the given item to the tree.  If an item in the tree
// already equals the given one, it is removed from the tree and returned,
// and the second return value is true.  Otherwise, (nil, false)
func (t *templateBTree) ReplaceOrInsert(pt *TemplateParser) (*TemplateParser, bool) {
	t.Lock()
	defer t.Unlock()

	item, ok := t.templates.ReplaceOrInsert(parsedTemplateBTreeItem{uid: pt.base.UID, t: pt})
	if !ok {
		return nil, false
	}
	return item.t, true
}

// Ascend calls the given function for each item in the tree in ascending order.
func (t *templateBTree) Ascend(fn func(*TemplateParser) bool) {
	t.RLock()
	defer t.RUnlock()

	t.templates.Ascend(func(elem parsedTemplateBTreeItem) bool {
		return fn(elem.t)
	})
}
