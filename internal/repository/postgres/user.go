// package postgres

// import (
// 	"errors"
// 	"todo-app/domain"
// 	"todo-app/pkg/clients"

// 	"gorm.io/gorm"
// )

// type userRepo struct {
// 	db *gorm.DB
// }

// func NewUserRepo(db *gorm.DB) *userRepo {
// 	return &userRepo{
// 		db: db,
// 	}
// }

// func (r *userRepo) Save(user *domain.UserCreate) error {
// 	if err := r.db.Create(&user).Error; err != nil {
// 		return clients.ErrDB(err)
// 	}

// 	return nil
// }

// func (r *userRepo) GetUser(conditions map[string]any) (*domain.User, error) {
// 	var user domain.User

// 	if err := r.db.Where(conditions).First(&user).Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, clients.ErrRecordNotFound
// 		}

// 		return nil, clients.ErrDB(err)
// 	}

// 	return &user, nil
// }

//////////////////////////////////////////////////////////////////////////////////////////////

package postgres

import (
	"errors"
	"todo-app/domain"
	"todo-app/pkg/clients"

	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *userRepo {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Save(user *domain.UserCreate) error {
	if err := r.db.Create(&user).Error; err != nil {
		return clients.ErrDB(err)
	}

	return nil
}

func (r *userRepo) GetAll(filter map[string]any, paging *clients.Paging) ([]domain.User, error) {
	users := []domain.User{}
	var query *gorm.DB

	if f := filter; f != nil {
		if v := f["user_id"]; v != "" {
			query = r.db.Where("user_id = ?", v)
		}
	}

	if err := query.Table(domain.User{}.TableName()).Select("id").Count(&paging.Total).Error; err != nil {
		return nil, clients.ErrDB(err)
	}

	query = r.db.Limit(paging.Limit).Offset((paging.Page - 1) * paging.Limit)

	if err := query.Find(&users).Error; err != nil {
		return nil, clients.ErrDB(err)
	}

	return users, nil
}

func (r *userRepo) GetUser(conditions map[string]any) (*domain.User, error) {
	var user domain.User

	if err := r.db.Where(conditions).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, clients.ErrRecordNotFound
		}

		return nil, clients.ErrDB(err)
	}

	return &user, nil
}

func (r *userRepo) Update(filter map[string]any, user *domain.UserUpdate) error {
	if err := r.db.Where(filter).Updates(&user).Error; err != nil {
		return clients.ErrDB(err)
	}

	return nil
}

func (r *userRepo) Delete(filter map[string]any) error {
	if err := r.db.Table(domain.User{}.TableName()).Where(filter).Delete(nil).Error; err != nil {
		return clients.ErrDB(err)
	}

	return nil
}
