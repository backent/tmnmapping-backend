package models

import "database/sql"

// --- Category ---

type Category struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NullAbleCategory struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

var CategoryTable string = "categories"

func NullAbleCategoryToCategory(n NullAbleCategory) Category {
	return Category{
		Id:        int(n.Id.Int64),
		Name:      n.Name.String,
		CreatedAt: n.CreatedAt.String,
		UpdatedAt: n.UpdatedAt.String,
	}
}

// --- SubCategory ---

type SubCategory struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NullAbleSubCategory struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

var SubCategoryTable string = "sub_categories"

func NullAbleSubCategoryToSubCategory(n NullAbleSubCategory) SubCategory {
	return SubCategory{
		Id:        int(n.Id.Int64),
		Name:      n.Name.String,
		CreatedAt: n.CreatedAt.String,
		UpdatedAt: n.UpdatedAt.String,
	}
}

// --- MotherBrand ---

type MotherBrand struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NullAbleMotherBrand struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

var MotherBrandTable string = "mother_brands"

func NullAbleMotherBrandToMotherBrand(n NullAbleMotherBrand) MotherBrand {
	return MotherBrand{
		Id:        int(n.Id.Int64),
		Name:      n.Name.String,
		CreatedAt: n.CreatedAt.String,
		UpdatedAt: n.UpdatedAt.String,
	}
}

// --- Branch ---

type Branch struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NullAbleBranch struct {
	Id        sql.NullInt64
	Name      sql.NullString
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

var BranchTable string = "branches"

func NullAbleBranchToBranch(n NullAbleBranch) Branch {
	return Branch{
		Id:        int(n.Id.Int64),
		Name:      n.Name.String,
		CreatedAt: n.CreatedAt.String,
		UpdatedAt: n.UpdatedAt.String,
	}
}
