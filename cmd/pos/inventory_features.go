package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"termpos/internal/auth"
	"termpos/internal/db"
	"termpos/internal/models"
)

var (
	// Flags for category command
	categoryDescription string
	categoryParentID    int

	// Flags for supplier command
	supplierContact string
	supplierEmail   string
	supplierPhone   string
	supplierAddress string
	supplierNotes   string

	// Flags for location command
	locationAddress     string
	locationDescription string

	// Flags for batch command
	batchLocationID      int
	batchSupplierID      int
	batchNumber          string
	batchExpiryDate      string
	batchManufactureDate string
	batchCostPrice       float64
	batchQuantity        int
	
	// Flags for product
	productCategory    int
	productSupplier    int
	productLowStock    int
	productSKU         string
	productDescription string
)

func init() {
	// Category commands
	categoryCmd := &cobra.Command{
		Use:   "category",
		Short: "Manage product categories",
		Long:  `Add, list, and manage product categories.`,
	}

	addCategoryCmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new product category",
		Long:  `Add a new product category with optional description and parent category.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAddCategory,
	}
	addCategoryCmd.Flags().StringVar(&categoryDescription, "description", "", "Description of the category")
	addCategoryCmd.Flags().IntVar(&categoryParentID, "parent", 0, "Parent category ID (0 for top-level category)")

	listCategoriesCmd := &cobra.Command{
		Use:   "list",
		Short: "List all product categories",
		Long:  `Display a list of all product categories.`,
		RunE:  runListCategories,
	}

	// Supplier commands
	supplierCmd := &cobra.Command{
		Use:   "supplier",
		Short: "Manage product suppliers",
		Long:  `Add, list, and manage product suppliers.`,
	}

	addSupplierCmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new supplier",
		Long:  `Add a new supplier with contact information.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAddSupplier,
	}
	addSupplierCmd.Flags().StringVar(&supplierContact, "contact", "", "Contact person name")
	addSupplierCmd.Flags().StringVar(&supplierEmail, "email", "", "Email address")
	addSupplierCmd.Flags().StringVar(&supplierPhone, "phone", "", "Phone number")
	addSupplierCmd.Flags().StringVar(&supplierAddress, "address", "", "Physical address")
	addSupplierCmd.Flags().StringVar(&supplierNotes, "notes", "", "Additional notes")

	listSuppliersCmd := &cobra.Command{
		Use:   "list",
		Short: "List all suppliers",
		Long:  `Display a list of all suppliers.`,
		RunE:  runListSuppliers,
	}

	// Location commands
	locationCmd := &cobra.Command{
		Use:   "location",
		Short: "Manage inventory locations",
		Long:  `Add, list, and manage inventory locations such as warehouses or stores.`,
	}

	addLocationCmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new location",
		Long:  `Add a new inventory location with optional address and description.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAddLocation,
	}
	addLocationCmd.Flags().StringVar(&locationAddress, "address", "", "Physical address")
	addLocationCmd.Flags().StringVar(&locationDescription, "description", "", "Description of the location")

	listLocationsCmd := &cobra.Command{
		Use:   "list",
		Short: "List all locations",
		Long:  `Display a list of all inventory locations.`,
		RunE:  runListLocations,
	}

	// Batch commands
	batchCmd := &cobra.Command{
		Use:   "batch",
		Short: "Manage product batches",
		Long:  `Add, list, and manage product batches with expiration dates.`,
	}

	addBatchCmd := &cobra.Command{
		Use:   "add [product_id]",
		Short: "Add a new product batch",
		Long:  `Add a new batch of products with expiration date and location information.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAddBatch,
	}
	addBatchCmd.Flags().IntVar(&batchLocationID, "location", 1, "Location ID")
	addBatchCmd.Flags().IntVar(&batchSupplierID, "supplier", 1, "Supplier ID")
	addBatchCmd.Flags().StringVar(&batchNumber, "batch-number", "", "Batch number or identifier")
	addBatchCmd.Flags().StringVar(&batchExpiryDate, "expiry", "", "Expiry date (YYYY-MM-DD)")
	addBatchCmd.Flags().StringVar(&batchManufactureDate, "manufacture-date", "", "Manufacture date (YYYY-MM-DD)")
	addBatchCmd.Flags().Float64Var(&batchCostPrice, "cost", 0, "Cost price per unit")
	addBatchCmd.Flags().IntVar(&batchQuantity, "quantity", 0, "Quantity in this batch")
	addBatchCmd.MarkFlagRequired("quantity")

	listBatchesCmd := &cobra.Command{
		Use:   "list [product_id]",
		Short: "List batches for a product",
		Long:  `Display a list of all batches for a specific product.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runListBatches,
	}

	expiredBatchesCmd := &cobra.Command{
		Use:   "expired",
		Short: "List expired product batches",
		Long:  `Display a list of all expired product batches.`,
		RunE:  runListExpiredBatches,
	}

	// Low stock alerts
	lowStockCmd := &cobra.Command{
		Use:   "low-stock",
		Short: "List products with low stock levels",
		Long:  `Display a list of products that have stock levels below their alert threshold.`,
		RunE:  runListLowStock,
	}

	// Enhance existing add product command
	rootCmd.PersistentFlags().IntVar(&productCategory, "category", 1, "Category ID for the product")
	rootCmd.PersistentFlags().IntVar(&productSupplier, "supplier", 1, "Default supplier ID for the product")
	rootCmd.PersistentFlags().IntVar(&productLowStock, "low-stock-alert", 0, "Low stock alert threshold (0 to disable)")
	rootCmd.PersistentFlags().StringVar(&productSKU, "sku", "", "Stock Keeping Unit (SKU) identifier")
	rootCmd.PersistentFlags().StringVar(&productDescription, "description", "", "Product description")

	// Add inventory by location command
	inventoryByLocationCmd := &cobra.Command{
		Use:   "location-inventory [location_id]",
		Short: "List inventory at a specific location",
		Long:  `Display inventory levels for all products at a specific location.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runInventoryByLocation,
	}

	// Add inventory by category command
	inventoryByCategoryCmd := &cobra.Command{
		Use:   "category-inventory [category_id]",
		Short: "List inventory for a specific category",
		Long:  `Display inventory levels for all products in a specific category.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runInventoryByCategory,
	}

	// Add subcommands to category command
	categoryCmd.AddCommand(addCategoryCmd, listCategoriesCmd)

	// Add subcommands to supplier command
	supplierCmd.AddCommand(addSupplierCmd, listSuppliersCmd)

	// Add subcommands to location command
	locationCmd.AddCommand(addLocationCmd, listLocationsCmd)

	// Add subcommands to batch command
	batchCmd.AddCommand(addBatchCmd, listBatchesCmd, expiredBatchesCmd)

	// Add commands to root
	rootCmd.AddCommand(categoryCmd, supplierCmd, locationCmd, batchCmd, lowStockCmd, inventoryByLocationCmd, inventoryByCategoryCmd)
}

// runAddCategory adds a new product category
func runAddCategory(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - only admin and manager can add categories
	if err := auth.RequirePermission("product:manage"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	name := args[0]
	category := models.Category{
		Name:        name,
		Description: categoryDescription,
		ParentID:    categoryParentID,
	}

	id, err := db.AddCategory(category)
	if err != nil {
		return fmt.Errorf("failed to add category: %v", err)
	}

	fmt.Printf("Category added successfully with ID: %d\n", id)
	return nil
}

// runListCategories lists all product categories
func runListCategories(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view categories
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	categories, err := db.GetAllCategories()
	if err != nil {
		return fmt.Errorf("failed to get categories: %v", err)
	}

	if len(categories) == 0 {
		fmt.Println("No categories found.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Description", "Parent ID"})

	for _, c := range categories {
		parentID := ""
		if c.ParentID > 0 {
			parentID = strconv.Itoa(c.ParentID)
		}

		table.Append([]string{
			strconv.Itoa(c.ID),
			c.Name,
			c.Description,
			parentID,
		})
	}

	table.Render()
	return nil
}

// runAddSupplier adds a new supplier
func runAddSupplier(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - only admin and manager can add suppliers
	if err := auth.RequirePermission("product:manage"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	name := args[0]
	supplier := models.Supplier{
		Name:     name,
		Contact:  supplierContact,
		Email:    supplierEmail,
		Phone:    supplierPhone,
		Address:  supplierAddress,
		Notes:    supplierNotes,
		IsActive: true,
	}

	id, err := db.AddSupplier(supplier)
	if err != nil {
		return fmt.Errorf("failed to add supplier: %v", err)
	}

	fmt.Printf("Supplier added successfully with ID: %d\n", id)
	return nil
}

// runListSuppliers lists all suppliers
func runListSuppliers(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view suppliers
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	suppliers, err := db.GetAllSuppliers()
	if err != nil {
		return fmt.Errorf("failed to get suppliers: %v", err)
	}

	if len(suppliers) == 0 {
		fmt.Println("No suppliers found.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Contact", "Email", "Phone", "Status"})

	for _, s := range suppliers {
		status := "Inactive"
		if s.IsActive {
			status = "Active"
		}

		table.Append([]string{
			strconv.Itoa(s.ID),
			s.Name,
			s.Contact,
			s.Email,
			s.Phone,
			status,
		})
	}

	table.Render()
	return nil
}

// runAddLocation adds a new inventory location
func runAddLocation(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - only admin and manager can add locations
	if err := auth.RequirePermission("product:manage"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	name := args[0]
	location := models.Location{
		Name:        name,
		Address:     locationAddress,
		Description: locationDescription,
		IsActive:    true,
	}

	id, err := db.AddLocation(location)
	if err != nil {
		return fmt.Errorf("failed to add location: %v", err)
	}

	fmt.Printf("Location added successfully with ID: %d\n", id)
	return nil
}

// runListLocations lists all inventory locations
func runListLocations(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view locations
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	locations, err := db.GetAllLocations()
	if err != nil {
		return fmt.Errorf("failed to get locations: %v", err)
	}

	if len(locations) == 0 {
		fmt.Println("No locations found.")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Address", "Description", "Status"})

	for _, l := range locations {
		status := "Inactive"
		if l.IsActive {
			status = "Active"
		}

		table.Append([]string{
			strconv.Itoa(l.ID),
			l.Name,
			l.Address,
			l.Description,
			status,
		})
	}

	table.Render()
	return nil
}

// runAddBatch adds a new product batch
func runAddBatch(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - only admin and manager can add product batches
	if err := auth.RequirePermission("product:manage"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	if batchQuantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}

	productID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid product ID: %v", err)
	}

	batch := models.ProductBatch{
		ProductID:    productID,
		LocationID:   batchLocationID,
		SupplierID:   batchSupplierID,
		Quantity:     batchQuantity,
		BatchNumber:  batchNumber,
		CostPrice:    batchCostPrice,
		ReceiptDate:  time.Now(),
	}

	// Parse expiry date if provided
	if batchExpiryDate != "" {
		expiryDate, err := time.Parse("2006-01-02", batchExpiryDate)
		if err != nil {
			return fmt.Errorf("invalid expiry date format (required: YYYY-MM-DD): %v", err)
		}
		batch.ExpiryDate = expiryDate
	}

	// Parse manufacture date if provided
	if batchManufactureDate != "" {
		manufactureDate, err := time.Parse("2006-01-02", batchManufactureDate)
		if err != nil {
			return fmt.Errorf("invalid manufacture date format (required: YYYY-MM-DD): %v", err)
		}
		batch.ManufactureDate = manufactureDate
	}

	id, err := db.AddProductBatch(batch)
	if err != nil {
		return fmt.Errorf("failed to add product batch: %v", err)
	}

	fmt.Printf("Product batch added successfully with ID: %d\n", id)
	product, err := db.GetProductByID(productID)
	if err != nil {
		return nil // We'll just skip showing the updated stock if there's an error
	}
	fmt.Printf("Product '%s' stock updated to: %d\n", product.Name, product.Stock)
	return nil
}

// runListBatches lists all batches for a product
func runListBatches(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view batches
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	productID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid product ID: %v", err)
	}

	// Get product details
	product, err := db.GetProductByID(productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %v", err)
	}

	batches, err := db.GetProductBatches(productID)
	if err != nil {
		return fmt.Errorf("failed to get product batches: %v", err)
	}

	if len(batches) == 0 {
		fmt.Printf("No batches found for product '%s' (ID: %d).\n", product.Name, product.ID)
		return nil
	}

	fmt.Printf("Batches for product: %s (ID: %d)\n\n", product.Name, product.ID)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Batch #", "Location", "Qty", "Expiry Date", "Cost Price", "Status"})

	for _, b := range batches {
		// Get location name
		location, err := db.GetLocationByID(b.LocationID)
		locationName := fmt.Sprintf("%d", b.LocationID)
		if err == nil {
			locationName = location.Name
		}

		// Format expiry date
		expiryDate := "N/A"
		if !b.ExpiryDate.IsZero() {
			expiryDate = b.ExpiryDate.Format("2006-01-02")
		}

		// Determine status
		status := "Good"
		if b.IsExpired() {
			status = "EXPIRED"
		} else if b.IsExpiringSoon(30) {
			status = "Expiring Soon"
		}

		table.Append([]string{
			strconv.Itoa(b.ID),
			b.BatchNumber,
			locationName,
			strconv.Itoa(b.Quantity),
			expiryDate,
			fmt.Sprintf("$%.2f", b.CostPrice),
			status,
		})
	}

	table.Render()
	return nil
}

// runListExpiredBatches lists all expired product batches
func runListExpiredBatches(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view expired batches
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	batches, err := db.GetExpiredBatches()
	if err != nil {
		return fmt.Errorf("failed to get expired batches: %v", err)
	}

	if len(batches) == 0 {
		fmt.Println("No expired batches found.")
		return nil
	}

	fmt.Println("Expired Product Batches:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Product", "Batch #", "Location", "Qty", "Expiry Date", "Days Expired"})

	for _, b := range batches {
		// Get product name
		product, err := db.GetProductByID(b.ProductID)
		productName := fmt.Sprintf("ID: %d", b.ProductID)
		if err == nil {
			productName = product.Name
		}

		// Get location name
		location, err := db.GetLocationByID(b.LocationID)
		locationName := fmt.Sprintf("ID: %d", b.LocationID)
		if err == nil {
			locationName = location.Name
		}

		// Calculate days expired
		daysExpired := time.Since(b.ExpiryDate).Hours() / 24

		table.Append([]string{
			strconv.Itoa(b.ID),
			productName,
			b.BatchNumber,
			locationName,
			strconv.Itoa(b.Quantity),
			b.ExpiryDate.Format("2006-01-02"),
			fmt.Sprintf("%.0f", daysExpired),
		})
	}

	table.Render()
	return nil
}

// runListLowStock lists all products with stock below the alert threshold
func runListLowStock(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view low stock
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	products, err := db.GetLowStockProducts()
	if err != nil {
		return fmt.Errorf("failed to get low stock products: %v", err)
	}

	if len(products) == 0 {
		fmt.Println("No products with low stock found.")
		return nil
	}

	fmt.Println("Low Stock Alert:")
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Category", "Current Stock", "Alert Threshold", "Supplier"})

	for _, p := range products {
		table.Append([]string{
			strconv.Itoa(p.ID),
			p.Name,
			p.CategoryName,
			strconv.Itoa(p.Stock),
			strconv.Itoa(p.LowStockAlert),
			p.SupplierName,
		})
	}

	table.SetRowLine(true)
	table.Render()
	return nil
}

// runInventoryByLocation lists inventory for a specific location
func runInventoryByLocation(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view inventory
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	locationID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid location ID: %v", err)
	}

	// Get location details
	location, err := db.GetLocationByID(locationID)
	if err != nil {
		return fmt.Errorf("failed to get location: %v", err)
	}

	products, err := db.GetProductsByLocation(locationID)
	if err != nil {
		return fmt.Errorf("failed to get products for location: %v", err)
	}

	if len(products) == 0 {
		fmt.Printf("No products found at location '%s' (ID: %d).\n", location.Name, location.ID)
		return nil
	}

	fmt.Printf("Inventory at location: %s (ID: %d)\n\n", location.Name, location.ID)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Category", "Quantity", "Batches", "Expiring"})

	for _, p := range products {
		expiring := "No"
		if p.HasExpiredBatches {
			expiring = "Yes"
		}

		table.Append([]string{
			strconv.Itoa(p.ID),
			p.Name,
			p.CategoryName,
			strconv.Itoa(p.Stock),
			strconv.Itoa(p.BatchCount),
			expiring,
		})
	}

	table.Render()
	return nil
}

// runInventoryByCategory lists inventory for a specific category
func runInventoryByCategory(cmd *cobra.Command, args []string) error {
	// Ensure user is authenticated
	if !auth.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	// Check permission - anyone with inventory:view can view inventory
	if err := auth.RequirePermission("inventory:view"); err != nil {
		return fmt.Errorf("unauthorized: %v", err)
	}

	categoryID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid category ID: %v", err)
	}

	// Get category details
	category, err := db.GetCategoryByID(categoryID)
	if err != nil {
		return fmt.Errorf("failed to get category: %v", err)
	}

	products, err := db.GetProductsByCategory(categoryID)
	if err != nil {
		return fmt.Errorf("failed to get products for category: %v", err)
	}

	if len(products) == 0 {
		fmt.Printf("No products found in category '%s' (ID: %d).\n", category.Name, category.ID)
		return nil
	}

	fmt.Printf("Inventory in category: %s (ID: %d)\n\n", category.Name, category.ID)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Price", "Stock", "Supplier", "Low Stock"})

	for _, p := range products {
		lowStock := "No"
		if p.IsLowStock {
			lowStock = "Yes"
		}

		table.Append([]string{
			strconv.Itoa(p.ID),
			p.Name,
			fmt.Sprintf("$%.2f", p.Price),
			strconv.Itoa(p.Stock),
			p.SupplierName,
			lowStock,
		})
	}

	table.Render()
	return nil
}