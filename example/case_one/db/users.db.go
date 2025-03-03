package db

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"math"
	"time"
)

// userStorage is a struct for the "users" table.
type userStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// UserCRUDOperations is an interface for managing the users table.
type UserCRUDOperations interface {
	Create(ctx context.Context, model *User, opts ...Option) (*string, error)
	BatchCreate(ctx context.Context, models []*User, opts ...Option) ([]string, error)
	Update(ctx context.Context, id string, updateData *UserUpdate) error
	DeleteById(ctx context.Context, id string, opts ...Option) error
	FindById(ctx context.Context, id string, opts ...Option) (*User, error)
}

// UserSearchOperations is an interface for searching the users table.
type UserSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*User, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*User, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*User, error)
}

// UserPaginationOperations is an interface for pagination operations.
type UserPaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*User, *Paginator, error)
}

// UserRelationLoading is an interface for loading relations.
type UserRelationLoading interface {
	LoadDevice(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadSettings(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadAddresses(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadPosts(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadBatchDevice(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchSettings(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchAddresses(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchPosts(ctx context.Context, items []*User, builders ...*QueryBuilder) error
}

// UserAdvancedDeletion is an interface for advanced deletion operations.
type UserAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// UserRawQueryOperations is an interface for executing raw queries.
type UserRawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// UserStorage is a struct for the "users" table.
type UserStorage interface {
	UserCRUDOperations
	UserSearchOperations
	UserPaginationOperations
	UserRelationLoading
	UserAdvancedDeletion
	UserRawQueryOperations
}

// NewUserStorage returns a new userStorage.
func NewUserStorage(config *Config) (UserStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB is nil")
	}
	if config.DB.DBRead == nil {
		return nil, errors.New("config.DB.DBRead is nil")
	}
	if config.DB.DBWrite == nil {
		config.DB.DBWrite = config.DB.DBRead
	}

	return &userStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *userStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *userStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// TableName returns the table name.
func (t *userStorage) TableName() string {
	return "users"
}

// Columns returns the columns for the table.
func (t *userStorage) Columns() []string {
	return []string{
		"id", "name", "age", "email", "last_name", "created_at", "updated_at", "notification_settings", "phones", "balls", "numrs", "comments",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *userStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
	var db QueryExecer

	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		if tx == nil {
			t.logError(ctx, errors.New("transaction is nil"), "failed to get transaction from context")
			// set default connection
			return t.config.DB.DBWrite
		}

		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		db = t.config.DB.DBWrite
	} else {
		db = t.config.DB.DBRead
	}

	return db
}

// LoadDevice loads the Device relation.
func (t *userStorage) LoadDevice(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewDeviceStorage creates a new DeviceStorage.
	s, err := NewDeviceStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create DeviceStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(DeviceUserIdEq(model.Id)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find one DeviceStorage")
	}

	model.Device = relationModel
	return nil
}

// LoadSettings loads the Settings relation.
func (t *userStorage) LoadSettings(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewSettingStorage creates a new SettingStorage.
	s, err := NewSettingStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create SettingStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(SettingUserIdEq(model.Id)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find one SettingStorage")
	}

	model.Settings = relationModel
	return nil
}

// LoadAddresses loads the Addresses relation.
func (t *userStorage) LoadAddresses(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewAddressStorage creates a new AddressStorage.
	s, err := NewAddressStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create AddressStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(AddressUserIdEq(model.Id)))
	relationModels, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many AddressStorage")
	}

	model.Addresses = relationModels
	return nil
}

// LoadPosts loads the Posts relation.
func (t *userStorage) LoadPosts(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewPostStorage creates a new PostStorage.
	s, err := NewPostStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create PostStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(PostAuthorIdEq(model.Id)))
	relationModels, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many PostStorage")
	}

	model.Posts = relationModels
	return nil
}

// LoadBatchDevice loads the Device relation.
func (t *userStorage) LoadBatchDevice(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewDeviceStorage creates a new DeviceStorage.
	s, err := NewDeviceStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create DeviceStorage")
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(DeviceUserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many DeviceStorage")
	}
	resultMap := make(map[interface{}]*Device)
	for _, result := range results {
		resultMap[result.UserId] = result
	}

	// Assign Device to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Device = v
		}
	}

	return nil
}

// LoadBatchSettings loads the Settings relation.
func (t *userStorage) LoadBatchSettings(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewSettingStorage creates a new SettingStorage.
	s, err := NewSettingStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create SettingStorage")
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(SettingUserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many SettingStorage")
	}
	resultMap := make(map[interface{}]*Setting)
	for _, result := range results {
		resultMap[result.UserId] = result
	}

	// Assign Setting to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Settings = v
		}
	}

	return nil
}

// LoadBatchAddresses loads the Addresses relation.
func (t *userStorage) LoadBatchAddresses(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewAddressStorage creates a new AddressStorage.
	s, err := NewAddressStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create AddressStorage")
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(AddressUserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many AddressStorage")
	}
	resultMap := make(map[interface{}][]*Address)
	for _, result := range results {
		resultMap[result.UserId] = append(resultMap[result.UserId], result)
	}

	// Assign Address to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Addresses = v
		}
	}

	return nil
}

// LoadBatchPosts loads the Posts relation.
func (t *userStorage) LoadBatchPosts(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewPostStorage creates a new PostStorage.
	s, err := NewPostStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create PostStorage")
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(PostAuthorIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many PostStorage")
	}
	resultMap := make(map[interface{}][]*Post)
	for _, result := range results {
		resultMap[result.AuthorId] = append(resultMap[result.AuthorId], result)
	}

	// Assign Post to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Posts = v
		}
	}

	return nil
}

// User is a struct for the "users" table.
type User struct {
	Id                   string  `db:"id"`
	Name                 string  `db:"name"`
	Age                  int32   `db:"age"`
	Email                string  `db:"email"`
	LastName             *string `db:"last_name"`
	Device               *Device
	Settings             *Setting
	Addresses            []*Address
	Posts                []*Post
	CreatedAt            time.Time                `db:"created_at"`
	UpdatedAt            *time.Time               `db:"updated_at"`
	NotificationSettings *UserNotificationSetting `db:"notification_settings"`
	Phones               UserPhonesRepeated       `db:"phones"`
	Balls                UserBallsRepeated        `db:"balls"`
	Numrs                UserNumrsRepeated        `db:"numrs"`
	Comments             UserCommentsRepeated     `db:"comments"`
}

// TableName returns the table name.
func (t *User) TableName() string {
	return "users"
}

// ScanRow scans a row into a User.
func (t *User) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.Name, &t.Age, &t.Email, &t.LastName, &t.CreatedAt, &t.UpdatedAt, &t.NotificationSettings, &t.Phones, &t.Balls, &t.Numrs, &t.Comments)
}

// ScanRows scans a single row into the User.
func (t *User) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.Name,
		&t.Age,
		&t.Email,
		&t.LastName,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.NotificationSettings,
		&t.Phones,
		&t.Balls,
		&t.Numrs,
		&t.Comments,
	)
}

// UserFilters is a struct that holds filters for User.
type UserFilters struct {
	Id    *string
	Name  *string
	Age   *int32
	Email *string
}

// UserIdEq returns a condition that checks if the field equals the value.
func UserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// UserNameEq returns a condition that checks if the field equals the value.
func UserNameEq(value string) FilterApplier {
	return EqualsCondition{Field: "name", Value: value}
}

// UserAgeEq returns a condition that checks if the field equals the value.
func UserAgeEq(value int32) FilterApplier {
	return EqualsCondition{Field: "age", Value: value}
}

// UserEmailEq returns a condition that checks if the field equals the value.
func UserEmailEq(value string) FilterApplier {
	return EqualsCondition{Field: "email", Value: value}
}

// UserIdNotEq returns a condition that checks if the field equals the value.
func UserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// UserNameNotEq returns a condition that checks if the field equals the value.
func UserNameNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "name", Value: value}
}

// UserAgeNotEq returns a condition that checks if the field equals the value.
func UserAgeNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "age", Value: value}
}

// UserEmailNotEq returns a condition that checks if the field equals the value.
func UserEmailNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "email", Value: value}
}

// UserIdGT greaterThanCondition than condition.
func UserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// UserNameGT greaterThanCondition than condition.
func UserNameGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "name", Value: value}
}

// UserAgeGT greaterThanCondition than condition.
func UserAgeGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "age", Value: value}
}

// UserEmailGT greaterThanCondition than condition.
func UserEmailGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "email", Value: value}
}

// UserIdLT less than condition.
func UserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// UserNameLT less than condition.
func UserNameLT(value string) FilterApplier {
	return LessThanCondition{Field: "name", Value: value}
}

// UserAgeLT less than condition.
func UserAgeLT(value int32) FilterApplier {
	return LessThanCondition{Field: "age", Value: value}
}

// UserEmailLT less than condition.
func UserEmailLT(value string) FilterApplier {
	return LessThanCondition{Field: "email", Value: value}
}

// UserIdGTE greater than or equal condition.
func UserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// UserNameGTE greater than or equal condition.
func UserNameGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "name", Value: value}
}

// UserAgeGTE greater than or equal condition.
func UserAgeGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "age", Value: value}
}

// UserEmailGTE greater than or equal condition.
func UserEmailGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "email", Value: value}
}

// UserIdLTE less than or equal condition.
func UserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// UserNameLTE less than or equal condition.
func UserNameLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "name", Value: value}
}

// UserAgeLTE less than or equal condition.
func UserAgeLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "age", Value: value}
}

// UserEmailLTE less than or equal condition.
func UserEmailLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "email", Value: value}
}

// UserIdBetween between condition.
func UserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// UserNameBetween between condition.
func UserNameBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "name", Min: min, Max: max}
}

// UserAgeBetween between condition.
func UserAgeBetween(min, max int32) FilterApplier {
	return BetweenCondition{Field: "age", Min: min, Max: max}
}

// UserEmailBetween between condition.
func UserEmailBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "email", Min: min, Max: max}
}

// UserIdILike iLike condition %
func UserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// UserNameILike iLike condition %
func UserNameILike(value string) FilterApplier {
	return ILikeCondition{Field: "name", Value: value}
}

// UserEmailILike iLike condition %
func UserEmailILike(value string) FilterApplier {
	return ILikeCondition{Field: "email", Value: value}
}

// UserIdLike like condition %
func UserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// UserNameLike like condition %
func UserNameLike(value string) FilterApplier {
	return LikeCondition{Field: "name", Value: value}
}

// UserEmailLike like condition %
func UserEmailLike(value string) FilterApplier {
	return LikeCondition{Field: "email", Value: value}
}

// UserIdNotLike not like condition
func UserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// UserNameNotLike not like condition
func UserNameNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "name", Value: value}
}

// UserEmailNotLike not like condition
func UserEmailNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "email", Value: value}
}

// UserIdIn condition
func UserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// UserNameIn condition
func UserNameIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "name", Values: values}
}

// UserAgeIn condition
func UserAgeIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "age", Values: values}
}

// UserEmailIn condition
func UserEmailIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "email", Values: values}
}

// UserIdNotIn not in condition
func UserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// UserNameNotIn not in condition
func UserNameNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "name", Values: values}
}

// UserAgeNotIn not in condition
func UserAgeNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "age", Values: values}
}

// UserEmailNotIn not in condition
func UserEmailNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "email", Values: values}
}

// UserIdOrderBy sorts the result in ascending order.
func UserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// UserNameOrderBy sorts the result in ascending order.
func UserNameOrderBy(asc bool) FilterApplier {
	return OrderBy("name", asc)
}

// UserAgeOrderBy sorts the result in ascending order.
func UserAgeOrderBy(asc bool) FilterApplier {
	return OrderBy("age", asc)
}

// UserEmailOrderBy sorts the result in ascending order.
func UserEmailOrderBy(asc bool) FilterApplier {
	return OrderBy("email", asc)
}

// Create creates a new User.
func (t *userStorage) Create(ctx context.Context, model *User, opts ...Option) (*string, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	// get value of phones
	phones, err := model.Phones.Value()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value of Phones")
	}
	// get value of balls
	balls, err := model.Balls.Value()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value of Balls")
	}
	// get value of numrs
	numrs, err := model.Numrs.Value()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value of Numrs")
	}
	// get value of comments
	comments, err := model.Comments.Value()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value of Comments")
	}

	query := t.queryBuilder.Insert("users").
		Columns(
			"name",
			"age",
			"email",
			"last_name",
			"created_at",
			"updated_at",
			"notification_settings",
			"phones",
			"balls",
			"numrs",
			"comments",
		).
		Values(
			model.Name,
			model.Age,
			model.Email,
			nullValue(model.LastName),
			model.CreatedAt,
			nullValue(model.UpdatedAt),
			nullValue(model.NotificationSettings),
			phones,
			balls,
			numrs,
			comments,
		)

	// add RETURNING "id" to query
	query = query.Suffix("RETURNING \"id\"")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	var id string
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, errors.Wrap(err, "failed to create User")
	}

	if options.relations && model.Device != nil {
		s, err := NewDeviceStorage(t.config)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Device")
		}

		model.Device.UserId = id
		err = s.Create(ctx, model.Device)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Device")
		}
	}
	if options.relations && model.Settings != nil {
		s, err := NewSettingStorage(t.config)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Settings")
		}

		model.Settings.UserId = id
		_, err = s.Create(ctx, model.Settings)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Settings")
		}
	}
	if options.relations && model.Addresses != nil {
		for _, item := range model.Addresses {
			item.UserId = id
			s, err := NewAddressStorage(t.config)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create Addresses")
			}

			_, err = s.Create(ctx, item)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create Addresses")
			}
		}
	}

	return &id, nil
}

// BatchCreate creates multiple User records in a single batch.
func (t *userStorage) BatchCreate(ctx context.Context, models []*User, opts ...Option) ([]string, error) {
	if len(models) == 0 {
		return nil, errors.New("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return nil, errors.New("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
		Columns(
			"name",
			"age",
			"email",
			"last_name",
			"created_at",
			"updated_at",
			"notification_settings",
			"phones",
			"balls",
			"numrs",
			"comments",
		)

	for _, model := range models {
		if model == nil {
			return nil, errors.New("one of the models is nil")
		}
		query = query.Values(
			model.Name,
			model.Age,
			model.Email,
			nullValue(model.LastName),
			model.CreatedAt,
			nullValue(model.UpdatedAt),
			nullValue(model.NotificationSettings),
			phones,
			balls,
			numrs,
			comments,
		)
	}

	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT (" + options.ignoreConflictField + ") DO NOTHING RETURNING \"id\"")
	} else {
		query = query.Suffix("RETURNING \"id\"")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}
		return nil, errors.Wrap(err, "failed to execute bulk insert")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var returnIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "failed to scan id")
		}
		returnIDs = append(returnIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows iteration error")
	}

	return returnIDs, nil
}

// UserUpdate is used to update an existing User.
type UserUpdate struct {
	// Use regular pointer types for non-optional fields
	Name *string
	// Use regular pointer types for non-optional fields
	Age *int32
	// Use regular pointer types for non-optional fields
	Email *string
	// Use null types for optional fields
	LastName null.String
	// Use regular pointer types for non-optional fields
	CreatedAt *time.Time
	// Use null types for optional fields
	UpdatedAt null.Time
	// Use null types for optional fields
	NotificationSettings NullableJSON[*UserNotificationSetting]
	// Use regular pointer types for non-optional fields
	Phones *UserPhonesRepeated
	// Use regular pointer types for non-optional fields
	Balls *UserBallsRepeated
	// Use regular pointer types for non-optional fields
	Numrs *UserNumrsRepeated
	// Use regular pointer types for non-optional fields
	Comments *UserCommentsRepeated
}

// Update updates an existing User based on non-nil fields.
func (t *userStorage) Update(ctx context.Context, id string, updateData *UserUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("users")
	// Handle fields that are not optional using a nil check
	if updateData.Name != nil {
		query = query.Set("name", *updateData.Name) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Age != nil {
		query = query.Set("age", *updateData.Age) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Email != nil {
		query = query.Set("email", *updateData.Email) // Dereference pointer value
	}
	// Handle fields that are optional and can be explicitly set to NULL
	if updateData.LastName.Valid {
		// Handle null.String specifically
		if updateData.LastName.String == "" {
			query = query.Set("last_name", nil) // Explicitly set NULL for empty string
		} else {
			query = query.Set("last_name", updateData.LastName.ValueOrZero())
		}
	}
	// Handle fields that are not optional using a nil check
	if updateData.CreatedAt != nil {
		query = query.Set("created_at", *updateData.CreatedAt) // Dereference pointer value
	}
	// Handle fields that are optional and can be explicitly set to NULL
	if updateData.UpdatedAt.Valid {
		// Handle null.Time specifically
		if updateData.UpdatedAt.Time.IsZero() {
			query = query.Set("updated_at", nil) // Explicitly set NULL if time is zero
		} else {
			query = query.Set("updated_at", updateData.UpdatedAt.Time)
		}
	}
	// Handle fields that are optional and can be explicitly set to NULL
	if updateData.NotificationSettings.Valid {
		if updateData.NotificationSettings.Data == nil {
			query = query.Set("notification_settings", nil) // Explicitly set NULL
		} else {
			query = query.Set("notification_settings", updateData.NotificationSettings.Data)
		}
	}
	// Handle fields that are not optional using a nil check
	if updateData.Phones != nil {
		query = query.Set("phones", *updateData.Phones) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Balls != nil {
		query = query.Set("balls", *updateData.Balls) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Numrs != nil {
		query = query.Set("numrs", *updateData.Numrs) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Comments != nil {
		query = query.Set("comments", *updateData.Comments) // Dereference pointer value
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to update User")
	}

	return nil
}

// DeleteById - deletes a User by its id.
func (t *userStorage) DeleteById(ctx context.Context, id string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("users").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete User")
	}

	return nil
}

// DeleteMany removes entries from the users table using the provided filters
func (t *userStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("users")

	var withFilter bool
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.ApplyDelete(query)
			withFilter = true
		}
	}

	if !withFilter {
		return errors.New("filters are required for delete operation")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete users")
	}

	return nil
}

// FindById retrieves a User by its id.
func (t *userStorage) FindById(ctx context.Context, id string, opts ...Option) (*User, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(UserIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one User: ")
	}

	return model, nil
}

// FindMany finds multiple User based on the provided options.
func (t *userStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*User, error) {
	// build query
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName())

	// set default options
	options := &Options{}

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)

		// apply pagination
		if builder.pagination != nil {
			if builder.pagination.limit != nil {
				query = query.Limit(*builder.pagination.limit)
			}
			if builder.pagination.offset != nil {
				query = query.Offset(*builder.pagination.offset)
			}
		}

		// apply sorting
		for _, option := range builder.sortOptions {
			query = option.Apply(query)
		}

		// apply options
		for _, o := range builder.options {
			o(options)
		}
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var results []*User
	for rows.Next() {
		model := &User{}
		if err := model.ScanRows(rows); err != nil {
			return nil, errors.Wrap(err, "failed to scan User")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
	}

	return results, nil
}

// FindOne finds a single User based on the provided options.
func (t *userStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*User, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne User")
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Count counts User based on the provided options.
func (t *userStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
	// build query
	query := t.queryBuilder.Select("COUNT(*)").From(t.TableName())

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "failed to scan count")
	}

	return count, nil
}

// FindManyWithPagination finds multiple User with pagination support.
func (t *userStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*User, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count User")
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Build the pagination object
	paginator := &Paginator{
		TotalCount: totalCount,
		Limit:      limit,
		Page:       page,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	}

	// Add pagination to query builder
	builders = append(builders, PaginateBuilder(uint64(limit), uint64(offset)))

	// Find records using FindMany
	records, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to find User")
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the User for the given ID.
func (t *userStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*User, error) {
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName()).Suffix("FOR UPDATE")

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model User
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, errors.Wrap(err, "failed to scan User")
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *userStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *userStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *userStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
