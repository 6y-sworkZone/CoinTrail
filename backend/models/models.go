package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Accounts  []Account      `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) HashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

type AccountType string

const (
	AccountTypeCash     AccountType = "cash"
	AccountTypeBank     AccountType = "bank"
	AccountTypeWeChat   AccountType = "wechat"
	AccountTypeAlipay   AccountType = "alipay"
	AccountTypeCredit   AccountType = "credit"
	AccountTypeInvest   AccountType = "invest"
	AccountTypeOther    AccountType = "other"
)

type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Type      AccountType    `gorm:"size:20;not null" json:"type"`
	Balance   float64        `gorm:"default:0;not null" json:"balance"`
	Currency  string         `gorm:"size:10;default:'CNY'" json:"currency"`
	Icon      string         `gorm:"size:50" json:"icon"`
	Note      string         `gorm:"size:255" json:"note"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
}

type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
)

type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Name      string         `gorm:"size:50;not null" json:"name"`
	Type      CategoryType   `gorm:"size:20;not null" json:"type"`
	Icon      string         `gorm:"size:50" json:"icon"`
	Color     string         `gorm:"size:20" json:"color"`
	Sort      int            `gorm:"default:0" json:"sort"`
	IsDefault bool           `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
}

type TransactionType string

const (
	TransactionTypeExpense TransactionType = "expense"
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	AccountID       uint           `gorm:"not null;index" json:"account_id"`
	TargetAccountID *uint          `gorm:"index" json:"target_account_id,omitempty"`
	CategoryID      *uint          `gorm:"index" json:"category_id,omitempty"`
	Type            TransactionType `gorm:"size:20;not null" json:"type"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Note            string         `gorm:"size:255" json:"note"`
	TransactionDate time.Time      `gorm:"not null;index" json:"transaction_date"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	User            User           `gorm:"foreignKey:UserID" json:"-"`
	Account         Account        `gorm:"foreignKey:AccountID" json:"-"`
	TargetAccount   *Account       `gorm:"foreignKey:TargetAccountID" json:"-"`
	Category        *Category      `gorm:"foreignKey:CategoryID" json:"-"`
}

type Budget struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	CategoryID uint           `gorm:"not null;index" json:"category_id"`
	Amount    float64        `gorm:"not null" json:"amount"`
	Month     string         `gorm:"size:7;not null;index" json:"month"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
	Category  Category       `gorm:"foreignKey:CategoryID" json:"-"`
}

func GetDefaultExpenseCategories() []Category {
	return []Category{
		{Name: "餐饮", Icon: "🍜", Color: "#FF6B6B", Sort: 1},
		{Name: "交通", Icon: "🚗", Color: "#4ECDC4", Sort: 2},
		{Name: "购物", Icon: "🛒", Color: "#45B7D1", Sort: 3},
		{Name: "娱乐", Icon: "🎮", Color: "#96CEB4", Sort: 4},
		{Name: "医疗", Icon: "💊", Color: "#FFEAA7", Sort: 5},
		{Name: "教育", Icon: "📚", Color: "#DDA0DD", Sort: 6},
		{Name: "居住", Icon: "🏠", Color: "#98D8C8", Sort: 7},
		{Name: "通讯", Icon: "📱", Color: "#F7DC6F", Sort: 8},
		{Name: "其他支出", Icon: "📝", Color: "#BDC3C7", Sort: 9},
	}
}

func GetDefaultIncomeCategories() []Category {
	return []Category{
		{Name: "工资", Icon: "💰", Color: "#2ECC71", Sort: 1},
		{Name: "奖金", Icon: "🎁", Color: "#3498DB", Sort: 2},
		{Name: "投资", Icon: "📈", Color: "#9B59B6", Sort: 3},
		{Name: "兼职", Icon: "💼", Color: "#E67E22", Sort: 4},
		{Name: "其他收入", Icon: "📝", Color: "#1ABC9C", Sort: 5},
	}
}
