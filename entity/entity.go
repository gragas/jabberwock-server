package entity

import (
	"fmt"
	"github.com/gragas/jabberwock-server/inventory"
	"strings"
)

type Health float32
type Energy float32
type Spirit float32

type Entity struct {
	Id                uint64
	Name              string
	Health, MaxHealth Health
	Energy, MaxEnergy Energy
	Spirit, MaxSpirit Spirit
	X, Y, Xv, Yv      float32
	Inventory         inventory.Inventory
	Equipped          [20]*inventory.Item
}

func (entity Entity) Equip(item inventory.Item, slotNumber int) bool {
	if item.SlotNumber != slotNumber {
		return false
	}
	if item.SlotNumber == RingSlot || item.SlotNumber == PiercingSlot {
		offset := 0
		for entity.Equipped[item.SlotNumber+offset] != nil && offset < 4 {
			offset++
		}
		if offset == 4 {
			return false
		}
		*entity.Equipped[item.SlotNumber+offset] = item
		return true
	}
	if entity.Equipped[item.SlotNumber] != nil {
		return false
	}
	*entity.Equipped[item.SlotNumber] = item
	return true
}

func (entity Entity) Unequip(slotNumber int) (*inventory.Item, bool) {
	if entity.Equipped[slotNumber] != nil {
		itemPtr := entity.Equipped[slotNumber]
		entity.Equipped[slotNumber] = nil
		return itemPtr, true
	}
	return nil, false
}

func (entity Entity) AddToInventory(itemStack inventory.ItemStack, position int) bool {
	return entity.Inventory.AddItemStack(itemStack, position)
}

func (entity Entity) RemoveFromInventory(position int) (*inventory.ItemStack, bool) {
	return entity.Inventory.RemoveItemStack(position)
}

// These values are used to access the equipped member
// of entities. E.g.,
// entity.Equipped[HeadSlot] = $SOME_ITEM
const (
	HeadSlot = iota
	LeftShoulderSlot
	RightShoulderSlot
	LeftArmSlot
	RightArmSlot
	LeftHandSlot
	RightHandSlot
	TorsoSlot
	LegsSlot
	LeftFootSlot
	RightFootSlot
	RingSlot
	PiercingSlot
)

func (entity Entity) String() string {
	var itemNames []string
	for _, item := range entity.Equipped {
		itemNames = append(itemNames, item.Name)
	}
	return fmt.Sprintf("Entity <%v>\n"+
		"name: %v\n"+
		"health: %v\n"+
		"maxHealth: %v\n"+
		"energy: %v\n"+
		"maxEnergy: %v\n"+
		"spirit: %v\n"+
		"maxSpirit: %v\n"+
		"x, y: %.2f, %.2f\n"+
		"xv, yv: %.2f, %.2f\n"+
		"Inventory: %v\n"+
		"Equipped: "+strings.Join(itemNames, ", "),
		entity.Id, entity.Name,
		entity.Health, entity.MaxHealth,
		entity.Energy, entity.MaxEnergy,
		entity.Spirit, entity.MaxSpirit,
		entity.X, entity.Y, entity.Xv, entity.Yv,
		entity.Inventory, entity.Equipped)
}
