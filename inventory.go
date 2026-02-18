package main

import (
	"errors"
	"sync"
)

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type Order struct {
	ID        string `json:"id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

type Inventory struct {
	mu       sync.RWMutex
	Products map[string]*Product
}

var store = &Inventory{
	Products: map[string]*Product{
		"1":  {ID: "1", Name: "Widget", Price: 19.99, Stock: 100},
		"2":  {ID: "2", Name: "Gadget", Price: 29.99, Stock: 50},
		"3":  {ID: "3", Name: "Doohickey", Price: 9.99, Stock: 200},
		"4":  {ID: "4", Name: "Thingamajig", Price: 14.99, Stock: 75},
		"5":  {ID: "5", Name: "Whatchamacallit", Price: 24.99, Stock: 30},
		"6":  {ID: "6", Name: "Gizmo", Price: 39.99, Stock: 25},
		"7":  {ID: "7", Name: "Contraption", Price: 49.99, Stock: 15},
		"8":  {ID: "8", Name: "Device", Price: 34.99, Stock: 40},
		"9":  {ID: "9", Name: "Apparatus", Price: 44.99, Stock: 20},
		"10": {ID: "10", Name: "Mechanism", Price: 59.99, Stock: 10},
		"11": {ID: "11", Name: "Tool", Price: 12.99, Stock: 150},
		"12": {ID: "12", Name: "Instrument", Price: 89.99, Stock: 5},
		"13": {ID: "13", Name: "Component", Price: 7.99, Stock: 300},
		"14": {ID: "14", Name: "Module", Price: 79.99, Stock: 12},
		"15": {ID: "15", Name: "Unit", Price: 22.99, Stock: 60},
		"16": {ID: "16", Name: "Assembly", Price: 99.99, Stock: 8},
		"17": {ID: "17", Name: "System", Price: 199.99, Stock: 3},
		"18": {ID: "18", Name: "Platform", Price: 149.99, Stock: 6},
		"19": {ID: "19", Name: "Framework", Price: 119.99, Stock: 9},
		"20": {ID: "20", Name: "Solution", Price: 299.99, Stock: 2},
	},
}

type OrderResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetProduct retrieves a product by ID safely
func (inventory *Inventory) GetProduct(id string) (*Product, error) {
	// 1. Thread safety
	inventory.mu.RLock()
	defer inventory.mu.RUnlock()

	// 2. Logic
	product, exists := inventory.Products[id]
	if !exists {
		return nil, errors.New("product not found")
	}

	// 3. Return a copy
	p := *product
	return &p, nil
}

// Process order attempts to fulfill an order and update stock safely
func (inventory *Inventory) ProcessOrder(order *Order) OrderResult {
	inventory.mu.Lock()
	defer inventory.mu.Unlock()

	product, exists := inventory.Products[order.ProductID]
	if !exists {
		return OrderResult{Success: false, Message: "Product not found"}
	}

	if product.Stock < order.Quantity {
		return OrderResult{Success: false, Message: "Insufficient stock"}
	}

	product.Stock -= order.Quantity
	order.Status = "fulfilled"
	return OrderResult{Success: true, Message: "Order fulfilled successfully"}
}
