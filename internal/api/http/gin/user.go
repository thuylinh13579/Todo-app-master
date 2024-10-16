// package gin

// import (
// 	"net/http"
// 	"todo-app/domain"
// 	"todo-app/pkg/clients"
// 	"todo-app/pkg/tokenprovider"

// 	"github.com/gin-gonic/gin"
// )

// type UserService interface {
// 	Register(data *domain.UserCreate) error
// 	Login(data *domain.UserLogin) (tokenprovider.Token, error)
// }

// type userHandler struct {
// 	userService UserService
// }

// func NewUserHandler(apiVersion *gin.RouterGroup, svc UserService) {
// 	userHandler := &userHandler{
// 		userService: svc,
// 	}

// 	users := apiVersion.Group("/users")
// 	users.POST("/register", userHandler.RegisterUserHandler)
// 	users.POST("/login", userHandler.LoginHandler)
// }

// func (h *userHandler) RegisterUserHandler(c *gin.Context) {
// 	var data domain.UserCreate

// 	if err := c.ShouldBind(&data); err != nil {
// 		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

// 		return
// 	}

// 	if err := h.userService.Register(&data); err != nil {
// 		c.JSON(http.StatusBadRequest, err)

// 		return
// 	}

// 	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(data.ID))
// }

// func (h *userHandler) LoginHandler(c *gin.Context) {
// 	var data domain.UserLogin

// 	if err := c.ShouldBind(&data); err != nil {
// 		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

// 		return
// 	}

// 	token, err := h.userService.Login(&data)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, err)

// 		return
// 	}

// 	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(token))
// }

//////////////////////////////////////////////////////////////////////////////


package gin

import (
	"net/http"
	"todo-app/domain"
	"todo-app/pkg/clients"
	"todo-app/pkg/tokenprovider"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Register(data *domain.UserCreate) error
	Login(data *domain.UserLogin) (tokenprovider.Token, error)
	CreateUser(user *domain.UserCreate) error
	GetAllUser(userID uuid.UUID, paging *clients.Paging) ([]domain.User, error)
	GetUserByID(id, userID uuid.UUID) (domain.User, error)
	UpdateUser(id, userID uuid.UUID, user *domain.UserUpdate) error
	DeleteUser(id, userID uuid.UUID) error
}

type userHandler struct {
	userService UserService
}

func NewUserHandler(apiVersion *gin.RouterGroup, svc UserService, middlewareAuth func(c *gin.Context), middlewareRateLimit func(c *gin.Context)) {
	userHandler := &userHandler{
		userService: svc,
	}

	// users := apiVersion.Group("/users")
	users := apiVersion.Group("/users", middlewareAuth)
	users.POST("/register", userHandler.RegisterUserHandler)
	users.POST("/login", userHandler.LoginHandler)
	users.POST("", userHandler.CreateUserHandler)
	users.GET("", middlewareRateLimit, userHandler.GetAllUserHandler)
	users.GET("/:id", userHandler.GetUserHandler)
	users.PATCH("/:id", userHandler.UpdateUserHandler)
	users.DELETE("/:id", userHandler.DeleteUserHandler)
}

func (h *userHandler) RegisterUserHandler(c *gin.Context) {
	var data domain.UserCreate

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	if err := h.userService.Register(&data); err != nil {
		c.JSON(http.StatusBadRequest, err)

		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(data.ID))
}

func (h *userHandler) LoginHandler(c *gin.Context) {
	var data domain.UserLogin

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	token, err := h.userService.Login(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)

		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(token))
}

func (h *userHandler) CreateUserHandler(c *gin.Context) {
	var user domain.UserCreate

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	requester := c.MustGet(clients.CurrentUser).(clients.Requester)
	user.ID = requester.GetUserID()
	if err := h.userService.CreateUser(&user); err != nil {
		c.JSON(http.StatusBadRequest, err)

		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(user.ID))
}

// GetAllUserHandler retrieves all users.
//
// @Summary      Get all users
// @Description  This endpoint retrieves a list of all users.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200  {object}  clients.SuccessRes  "List of users retrieved successfully"
// @Failure      500  {object}  clients.AppError    "Internal Server Error"
// @Router       /users [get]
func (h *userHandler) GetAllUserHandler(c *gin.Context) {
	var paging clients.Paging
	if err := c.ShouldBind(&paging); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}
	paging.Process()

	requester := c.MustGet(clients.CurrentUser).(clients.Requester)

	users, err := h.userService.GetAllUser(requester.GetUserID(), &paging)
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	c.JSON(http.StatusOK, clients.NewSuccessResponse(users, paging, nil))
}

// GetItemHandler retrieves an item by its ID.
//
// @Summary      Get an user by ID
// @Description  This endpoint retrieves a single item by its unique identifier.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "User ID"
// @Success      200  {object}  clients.SuccessRes     "User retrieved successfully"
// @Failure      400  {object}  clients.AppError       "Invalid ID format or bad request"
// @Failure      404  {object}  clients.AppError       "User not found"
// @Failure      500  {object}  clients.AppError       "Internal Server Error"
// @Router       /users/{id} [get]
func (h *userHandler) GetUserHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	requester := c.MustGet(clients.CurrentUser).(clients.Requester)

	user, err := h.userService.GetUserByID(id, requester.GetUserID())
	if err != nil {
		c.JSON(http.StatusBadRequest, err)

		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(user))
}

// UpdateItemHandler updates an existing item.
//
// @Summary      Update an item
// @Description  This endpoint allows updating the properties of an existing item by its ID.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "User ID"
// @Param        item  body      domain.ItemUpdate      true  "Item update payload"
// @Success      200   {object}  clients.SuccessRes     "Item updated successfully"
// @Failure      400   {object}  clients.AppError       "Invalid input or bad request"
// @Failure      404   {object}  clients.AppError       "User not found"
// @Failure      500   {object}  clients.AppError       "Internal Server Error"
// @Router       /users/{id} [put]
func (h *userHandler) UpdateUserHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	user := domain.UserUpdate{}
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	requester := c.MustGet(clients.CurrentUser).(clients.Requester)

	if err := h.userService.UpdateUser(id, requester.GetUserID(), &user); err != nil {
		c.JSON(http.StatusBadRequest, err)

		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(true))
}

// DeleteItemHandler deletes an item by its ID.
//
// @Summary      Delete an user
// @Description  This endpoint deletes an user identified by its unique ID.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "User ID"
// @Success      200  {object}  clients.SuccessRes     "user deleted successfully"
// @Failure      400  {object}  clients.AppError       "Invalid ID format or bad request"
// @Failure      404  {object}  clients.AppError       "User not found"
// @Failure      500  {object}  clients.AppError       "Internal Server Error"
// @Router       /users/{id} [delete]
func (h *userHandler) DeleteUserHandler(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, clients.ErrInvalidRequest(err))

		return
	}

	requester := c.MustGet(clients.CurrentUser).(clients.Requester)

	if err := h.userService.DeleteUser(id, requester.GetUserID()); err != nil {
		c.JSON(http.StatusBadRequest, err)

		return
	}

	c.JSON(http.StatusOK, clients.SimpleSuccessResponse(true))
}
