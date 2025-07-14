package db

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    FirstName        string    `gorm:"not null"`
    LastName         string    `gorm:"not null"`
    TelegramUsername string    `gorm:"unique;not null"`
    TelegramID       *int64    `gorm:"unique"`
    Login            *string   `gorm:"unique"`
    PasswordHash     *string
    IsRegistered     bool      `gorm:"default:false"`
    IsAdmin          bool      `gorm:"default:false"`
    CreatedAt        time.Time
    RegisteredAt     *time.Time
}

type Transaction struct {
    ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    User      User      `gorm:"foreignKey:UserID"`
    Amount    int       `gorm:"not null"`
    Reason    string    `gorm:"not null"`
    Source    string    `gorm:"not null;check:source IN ('reward','purchase','adjustment')"`
    CreatedAt time.Time
}

type Product struct {
    ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Name        string    `gorm:"not null"`
    Description string
    Price       int       `gorm:"not null"`
    Stock       int       `gorm:"not null"`
    Type        string    `gorm:"not null;check:type IN ('regular','drop','shared')"`
    MediaURL    *string
    IsActive    bool      `gorm:"default:true"`
    CreatedAt   time.Time
}

type Purchase struct {
    ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    User      User      `gorm:"foreignKey:UserID"`
    ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
    Product   Product   `gorm:"foreignKey:ProductID"`
    Quantity  int       `gorm:"not null"`
    Confirmed bool      `gorm:"default:true"`
    CreatedAt time.Time
}

type LotParticipant struct {
    ID                uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    ProductID         uuid.UUID `gorm:"type:uuid;not null;index"`
    Product           Product   `gorm:"foreignKey:ProductID"`
    UserID            uuid.UUID `gorm:"type:uuid;not null;index"`
    User              User      `gorm:"foreignKey:UserID"`
    ContributedAmount int       `gorm:"not null"`
    CreatedAt         time.Time
}

type Banner struct {
    ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    Title     string    `gorm:"not null"`
    ImageURL  string    `gorm:"not null"`
    CreatedAt time.Time
}

type Feedback struct {
    ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    User      User      `gorm:"foreignKey:UserID"`
    Message   string    `gorm:"not null"`
    CreatedAt time.Time
}

type AdminLog struct {
    ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    AdminID      uuid.UUID `gorm:"type:uuid;not null;index"`
    Admin        User      `gorm:"foreignKey:AdminID"`
    TargetUserID uuid.UUID `gorm:"type:uuid;not null;index"`
    TargetUser   User      `gorm:"foreignKey:TargetUserID"`
    Action       string    `gorm:"not null;check:action IN ('reward','message','adjustment')"`
    Amount       *int
    Context      *string
    CreatedAt    time.Time
}

type MediaUpload struct {
    ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    UploaderID uuid.UUID `gorm:"type:uuid;not null;index"`
    Uploader   User      `gorm:"foreignKey:UploaderID"`
    FileName   string    `gorm:"not null"`
    FileType   string    `gorm:"not null;check:file_type IN ('image','video','gif','zip','other')"`
    URL        string    `gorm:"not null"`
    UploadedAt time.Time
}

type Notification struct {
    ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
    UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
    User      User       `gorm:"foreignKey:UserID"`
    Message   string     `gorm:"not null"`
    IsRead    bool       `gorm:"default:false"`
    SentAt    time.Time
    ReadAt    *time.Time
}
