package db

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type CommonRepo struct {
	db      orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// NewCommonRepo returns new repository
func NewCommonRepo(db orm.DB) CommonRepo {
	return CommonRepo{
		db: db,
		filters: map[string][]Filter{
			Tables.User.Name: {StatusFilter},
		},
		sort: map[string][]SortField{
			Tables.Stat.Name: {{Column: Columns.Stat.ID, Direction: SortDesc}},
			Tables.User.Name: {{Column: Columns.User.ID, Direction: SortDesc}},
		},
		join: map[string][]string{
			Tables.Stat.Name: {TableColumns, Columns.Stat.User},
			Tables.User.Name: {TableColumns},
		},
	}
}

// WithTransaction is a function that wraps CommonRepo with pg.Tx transaction.
func (cr CommonRepo) WithTransaction(tx *pg.Tx) CommonRepo {
	cr.db = tx
	return cr
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func (cr CommonRepo) WithEnabledOnly() CommonRepo {
	f := make(map[string][]Filter, len(cr.filters))
	for i := range cr.filters {
		f[i] = make([]Filter, len(cr.filters[i]))
		copy(f[i], cr.filters[i])
		f[i] = append(f[i], StatusEnabledFilter)
	}
	cr.filters = f

	return cr
}

/*** Stat ***/

// FullStat returns full joins with all columns
func (cr CommonRepo) FullStat() OpFunc {
	return WithColumns(cr.join[Tables.Stat.Name]...)
}

// DefaultStatSort returns default sort.
func (cr CommonRepo) DefaultStatSort() OpFunc {
	return WithSort(cr.sort[Tables.Stat.Name]...)
}

// StatByID is a function that returns Stat by ID(s) or nil.
func (cr CommonRepo) StatByID(ctx context.Context, id int64, ops ...OpFunc) (*Stat, error) {
	return cr.OneStat(ctx, &StatSearch{ID: &id}, ops...)
}

// OneStat is a function that returns one Stat by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneStat(ctx context.Context, search *StatSearch, ops ...OpFunc) (*Stat, error) {
	obj := &Stat{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.Stat.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// StatsByFilters returns Stat list.
func (cr CommonRepo) StatsByFilters(ctx context.Context, search *StatSearch, pager Pager, ops ...OpFunc) (stats []Stat, err error) {
	err = buildQuery(ctx, cr.db, &stats, search, cr.filters[Tables.Stat.Name], pager, ops...).Select()
	return
}

// CountStats returns count
func (cr CommonRepo) CountStats(ctx context.Context, search *StatSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &Stat{}, search, cr.filters[Tables.Stat.Name], PagerOne, ops...).Count()
}

// AddStat adds Stat to DB.
func (cr CommonRepo) AddStat(ctx context.Context, stat *Stat, ops ...OpFunc) (*Stat, error) {
	q := cr.db.ModelContext(ctx, stat)
	applyOps(q, ops...)
	_, err := q.Insert()

	return stat, err
}

// UpdateStat updates Stat in DB.
func (cr CommonRepo) UpdateStat(ctx context.Context, stat *Stat, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, stat).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.Stat.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteStat deletes Stat from DB.
func (cr CommonRepo) DeleteStat(ctx context.Context, id int64) (deleted bool, err error) {
	stat := &Stat{ID: id}

	res, err := cr.db.ModelContext(ctx, stat).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

/*** User ***/

// FullUser returns full joins with all columns
func (cr CommonRepo) FullUser() OpFunc {
	return WithColumns(cr.join[Tables.User.Name]...)
}

// DefaultUserSort returns default sort.
func (cr CommonRepo) DefaultUserSort() OpFunc {
	return WithSort(cr.sort[Tables.User.Name]...)
}

// UserByID is a function that returns User by ID(s) or nil.
func (cr CommonRepo) UserByID(ctx context.Context, id int, ops ...OpFunc) (*User, error) {
	return cr.OneUser(ctx, &UserSearch{ID: &id}, ops...)
}

// OneUser is a function that returns one User by filters. It could return pg.ErrMultiRows.
func (cr CommonRepo) OneUser(ctx context.Context, search *UserSearch, ops ...OpFunc) (*User, error) {
	obj := &User{}
	err := buildQuery(ctx, cr.db, obj, search, cr.filters[Tables.User.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// UsersByFilters returns User list.
func (cr CommonRepo) UsersByFilters(ctx context.Context, search *UserSearch, pager Pager, ops ...OpFunc) (users []User, err error) {
	err = buildQuery(ctx, cr.db, &users, search, cr.filters[Tables.User.Name], pager, ops...).Select()
	return
}

// CountUsers returns count
func (cr CommonRepo) CountUsers(ctx context.Context, search *UserSearch, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, cr.db, &User{}, search, cr.filters[Tables.User.Name], PagerOne, ops...).Count()
}

// AddUser adds User to DB.
func (cr CommonRepo) AddUser(ctx context.Context, user *User, ops ...OpFunc) (*User, error) {
	q := cr.db.ModelContext(ctx, user)
	applyOps(q, ops...)
	_, err := q.Insert()

	return user, err
}

// UpdateUser updates User in DB.
func (cr CommonRepo) UpdateUser(ctx context.Context, user *User, ops ...OpFunc) (bool, error) {
	q := cr.db.ModelContext(ctx, user).WherePK()
	if len(ops) == 0 {
		q = q.ExcludeColumn(Columns.User.ID)
	}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}

// DeleteUser set statusId to deleted in DB.
func (cr CommonRepo) DeleteUser(ctx context.Context, id int) (deleted bool, err error) {
	user := &User{ID: id, StatusID: StatusDisabled}

	return cr.UpdateUser(ctx, user, WithColumns(Columns.User.StatusID))
}
