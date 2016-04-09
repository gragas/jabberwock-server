package inventory

import (
	"fmt"
	"strings"
)

type Quantity uint32

type Item struct {
	Name                      string
	Durability, MaxDurability uint32
	SlotNumber                int
}

func (item Item) String() string {
	return fmt.Sprintf("%s", item.Name)
}

type ItemStack struct {
	Item     Item
	Quantity Quantity
}

func (itemStack ItemStack) String() string {
	return fmt.Sprintf("%v x %v", itemStack.Item, itemStack.Quantity)
}

const maxInventoryItemStacks uint32 = 32

type Inventory struct {
	ItemStacks [maxInventoryItemStacks]*ItemStack
}

func (inventory Inventory) String() string {
	var itemStackNames []string
	for _, itemStack := range inventory.ItemStacks {
		itemStackNames = append(itemStackNames, itemStack.Item.Name)
	}
	return strings.Join(itemStackNames, ", ")
}

func (inventory Inventory) AddItemStack(itemStack ItemStack, position int) bool {
	if inventory.ItemStacks[position] == nil {
		*inventory.ItemStacks[position] = itemStack
		return true
	} else if inventory.ItemStacks[position].Item.Name == itemStack.Item.Name {
		(*inventory.ItemStacks[position]).Quantity += itemStack.Quantity
		return true
	}
	return false
}

func (inventory Inventory) RemoveItemStack(position int) (*ItemStack, bool) {
	if inventory.ItemStacks[position] != nil {
		itemStackPtr := inventory.ItemStacks[position]
		inventory.ItemStacks[position] = nil
		return itemStackPtr, true
	}
	return nil, false
}
