package entity

import (
	"fmt"
	"inventory"
)

type health float
type energy float
type spirit float

type Entity struct {
	id                uint64
	name              string
	health, maxHealth health
	energy, maxEnergy energy
	spirit, maxSpirit spirit
	x, y, xv, yv      float
	inventory         inventory.Inventory
	equipped          [20]inventory.Item
}

func (entity Entity) AddToInventory(itemStack inventory.ItemStack) error {
	return entity.inventory.AddItemStack(itemStack) // either nil or some error
}

func (entity Entity) RemoveFromInventory(itemStack inventory.ItemStack) (inventory.ItemStack, error) {
	return entity.inventory.RemoveFromInventory(itemStack) // either nil or some error
}

// These values are used to access the equipped member
// of entities. E.g.,
// entity.equipped[head] = $SOME_ITEM
const (
	head = iota
	leftShoulder
	rightShoulder
	leftArm
	rightArm
	leftHand
	rightHand
	torso
	legs
	leftFoot
	rightFoot
	ring1
	ring2
	ring3
	ring4
	piercing1
	piercing2
	piercing3
	piercing4
)

func (entity Entity) String() string {
	var itemNames []string
	for _, item := range entity.equipped {
		append(itemNames, item.name)
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
		entity.id, entity.name,
		entity.health, entity.maxHealth,
		entity.energy, entity.maxEnergy,
		entity.spirit, entity.maxSpirit,
		entity.x, entity.y, entity.xv, entity.yv,
		entity.inventory, entity.equipped)
}
