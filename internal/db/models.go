package db

import (
    "time"
)

type Group struct {
    ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
    Name string `gorm:"not null" json:"name"`
}

type GroupAdmin struct {
    ID      uint `gorm:"primaryKey;autoIncrement"`
    GroupID uint
    AdminID uint // user_id администратора
}

type User struct {
    ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    FirstName        string    `gorm:"not null" json:"first_name"`
    LastName         string    `gorm:"not null" json:"last_name"`
    TelegramUsername string    `gorm:"unique;not null" json:"telegram_username"`
    TelegramID       *int64    `gorm:"unique" json:"telegram_id"`
    Login            *string   `gorm:"unique" json:"login"`
    PasswordHash     *string   `json:"password_hash"`
    IsRegistered     bool      `gorm:"default:false" json:"is_registered"`
    IsAdmin          bool      `gorm:"default:false" json:"is_admin"`
    Balance          int       `gorm:"default:0;not null" json:"balance"`
    RegisteredAt     *time.Time `json:"registered_at"`
    IsActive         bool      `gorm:"default:true" json:"is_active"`
    IsBanned         bool      `gorm:"default:false" json:"is_banned"`
    GroupID          *uint     `json:"group_id"`
    Group            *Group    `gorm:"foreignKey:GroupID" json:"group"`
}

type BulkUser struct {
    FirstName        string `json:"first_name"`
    LastName         string `json:"last_name"`
    TelegramUsername string `json:"telegram_username"`
    GroupID          *uint  `json:"group_id"`
    IsAdmin          bool   `json:"is_admin"`
}

type Transaction struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    User      User      `gorm:"foreignKey:UserID" json:"user"`
    Amount    int       `gorm:"not null" json:"amount"`
    Reason    string    `gorm:"not null" json:"reason"`
    Source    string    `gorm:"not null;check:source IN ('reward','purchase','adjustment')" json:"source"`
    CreatedAt time.Time `json:"created_at"`
}

type Product struct {
    ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"not null" json:"name"`
    Description string    `json:"description"`
    Price       int       `gorm:"not null" json:"price"`
    Stock       int       `gorm:"not null" json:"stock"`
    Type        string    `gorm:"not null;check:type IN ('regular','drop','shared')" json:"type"`
    MediaURL    *string   `json:"media_url"`
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
}

type Purchase struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    User      User      `gorm:"foreignKey:UserID" json:"user"`
    ProductID uint      `gorm:"not null;index" json:"product_id"`
    Product   Product   `gorm:"foreignKey:ProductID" json:"product"`
    Quantity  int       `gorm:"not null" json:"quantity"`
    Confirmed bool      `gorm:"default:true" json:"confirmed"`
    CreatedAt time.Time `json:"created_at"`
}

type LotParticipant struct {
    ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    ProductID         uint      `gorm:"not null;index" json:"product_id"`
    Product           Product   `gorm:"foreignKey:ProductID" json:"product"`
    UserID            uint      `gorm:"not null;index" json:"user_id"`
    User              User      `gorm:"foreignKey:UserID" json:"user"`
    ContributedAmount int       `gorm:"not null" json:"contributed_amount"`
    CreatedAt         time.Time `json:"created_at"`
}

type Banner struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Title     string    `gorm:"not null" json:"title"`
    ImageURL  string    `gorm:"not null" json:"image_url"`
    IsActive  bool      `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

type AdminLog struct {
    ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    AdminID      uint      `gorm:"not null;index" json:"admin_id"`
    Admin        User      `gorm:"foreignKey:AdminID" json:"admin"`
    TargetUserID uint      `gorm:"not null;index" json:"target_user_id"`
    TargetUser   User      `gorm:"foreignKey:TargetUserID" json:"target_user"`
    Action       string    `gorm:"not null;check:action IN ('reward','message','adjustment')" json:"action"`
    Amount       *int      `json:"amount"`
    Context      *string   `json:"context"`
    CreatedAt    time.Time `json:"created_at"`
}

type MediaUpload struct {
    ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    UploaderID uint      `gorm:"not null;index" json:"uploader_id"`
    Uploader   User      `gorm:"foreignKey:UploaderID" json:"uploader"`
    FileName   string    `gorm:"not null" json:"file_name"`
    FileType   string    `gorm:"not null;check:file_type IN ('image','video','gif','zip','other')" json:"file_type"`
    URL        string    `gorm:"not null" json:"url"`
    UploadedAt time.Time `json:"uploaded_at"`
}

type Notification struct {
    ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    uint       `gorm:"not null;index" json:"user_id"`
    User      User       `gorm:"foreignKey:UserID" json:"user"`
    Message   string     `gorm:"not null" json:"message"`
    IsRead    bool       `gorm:"default:false" json:"is_read"`
    SentAt    time.Time  `json:"sent_at"`
    ReadAt    *time.Time `json:"read_at"`
}

type BroadcastMessage struct {
    ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    AdminID           uint      `gorm:"not null;index" json:"admin_id"`
    Admin             User      `gorm:"foreignKey:AdminID" json:"admin"`
    Text              string    `gorm:"not null" json:"text"`
    MediaURL          *string   `json:"media_url"`
    MediaType         *string   `gorm:"check:media_type IN ('image','video','gif','document')" json:"media_type"`
    TargetType        string    `gorm:"not null;check:target_type IN ('all','registered','manual')" json:"target_type"`
    TargetUserIDs     string    `gorm:"type:text" json:"target_user_ids"` // JSON string для хранения массива ID
    SentCount         int       `gorm:"default:0" json:"sent_count"`
    FailedCount       int       `gorm:"default:0" json:"failed_count"`
    Status            string    `gorm:"not null;default:'pending';check:status IN ('pending','sending','completed','failed')" json:"status"`
    CreatedAt         time.Time `json:"created_at"`
    CompletedAt       *time.Time `json:"completed_at"`
}

type BroadcastHistory struct {
    ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    BroadcastID       uint      `gorm:"not null;index" json:"broadcast_id"`
    Broadcast         BroadcastMessage `gorm:"foreignKey:BroadcastID" json:"broadcast"`
    UserID            uint      `gorm:"not null;index" json:"user_id"`
    User              User      `gorm:"foreignKey:UserID" json:"user"`
    TelegramID        int64     `gorm:"not null" json:"telegram_id"`
    Status            string    `gorm:"not null;check:status IN ('sent','failed')" json:"status"`
    ErrorMessage      *string   `json:"error_message"`
    SentAt            time.Time `json:"sent_at"`
}

type Setting struct {
    Key   string `gorm:"primaryKey;size:64" json:"key"`
    Value string `gorm:"not null" json:"value"`
}
