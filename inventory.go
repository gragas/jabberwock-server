package inventory

import (
	"fmt"
	"string"
)

type Quantity uint32

type Item struct {
	name                      string
	durability, maxDurability uint32
}

func (item Item) String() string {
	return fmt.Sprintf("%s", item.name)
}

type ItemStack struct {
	item     Item
	quantity Quantity
}

func (itemStack ItemStack) String() string {
	return fmt.Sprintf("%v x %v", itemStack.item, itemStack.quantity)
}

const maxInventoryItemStacks uint32 = 32

type Inventory struct {
	itemStacks [maxInventoryItemStacks]ItemStack
}

func (inventory Inventory) String() string {
	var itemStackNames []string
	for _, itemStack := range inventory.itemStacks {
		append(itemStackNames, itemStack.item.name)
	}
	return strings.Join(itemStackNames, ", ")
}

func (inventory Inventory) AddItemStack(itemStack ItemStack, position int) error {
	if inventory.itemStacks[position] == nil {
		inventory.itemStacks[position] = itemStack
		return nil
	} else if inventory.itemStacks[position].item.name == itemStack.item.name {
		inventory.itemStacks[position].quantity += itemStack.quantity
		return nil
	}
	return errors.New(fmt.Sprintf(
		"Cannot add itemStack %v to inventory at position %v", item, position))
}

func (inventory Inventory) RemoveItemStack(position int) (ItemStack, error) {
	if inventory.itemStacks[position] != nil {
		itemStack := inventory.itemStacks[position]
		inventory.itemStacks[position] = nil
		return itemStack, nil
	}
	return nil, errors.New(fmt.Sprintf(
		"Cannot remove itemStack from inventory at position %v", position))
}
