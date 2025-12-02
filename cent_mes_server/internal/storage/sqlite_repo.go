package sqlite_repo

import (
	"cent_mes_server/models"
	"context"
	"database/sql"
	"strings"
	"time"
)

type SQLiteRepo struct {
	DB *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{DB: db}
}

func (r *SQLiteRepo) CreateTables(ctx context.Context) error {
	var err error
	_, err = r.DB.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS users_auth(
		login TEXT PRIMARY KEY,
		passwd TEXT NOT NULL,
		token TEXT NOT NULL);`)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS chats(
		id INTEGER PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT,
		CREATED_AT DATETIME NOT NULL);`)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS chat_members(
		chat_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		PRIMARY KEY (chat_id, user_id));`)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS messages(
		id INTEGER PRIMARY KEY,
		chat_id INTEGER NOT NULL,
		sender_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		body TEXT NOT NULL);`)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLiteRepo) Close() error {
	return r.DB.Close()
}
func (r *SQLiteRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	res := r.DB.QueryRowContext(ctx,
		`SELECT login, passwd, token FROM users_auth WHERE login = ?;`, login)
	var user models.User
	err := res.Scan(&user.Login, &user.Password, &user.Token)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return &user, err
}

func (r *SQLiteRepo) GetUserByToken(ctx context.Context, token string) (*models.User, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT * FROM users_auth WHERE token=?`, token)
	var user models.User
	err := row.Scan(&user.Login, &user.Password, &user.Token)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLiteRepo) RegisterUser(ctx context.Context, user *models.User) error {
	res, err := r.DB.ExecContext(ctx,
		`INSERT INTO users_auth (login, passwd, token) VALUES (?, ?, ?);`,
		user.Login, user.Password, user.Token)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func (r *SQLiteRepo) MessagesInChat(ctx context.Context, chatCF *models.ChatFetch) ([]models.Message, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT * FROM messages
	WHERE chat_id=? AND id>?
	ORDER BY created_at`,
		chatCF.ChatID, chatCF.AfterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		rows.Scan(&msg.ID, &msg.ChatID, &msg.Sender, &msg.CreatedAt, &msg.Body)
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *SQLiteRepo) UserChats(ctx context.Context, user *models.User) ([]models.Chat, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT 
		cm.chat_id, 
		c.title,
		(SELECT m.created_at
		FROM messages m
		WHERE m.chat_id=cm.chat_id
		ORDER BY m.created_at DESC
		LIMIT 1) as last_mes_at
		FROM chat_members cm
		JOIN chats c ON cm.chat_id=c.id
		WHERE user_id=?
		ORDER BY last_mes_at DESC`, user.Login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chats []models.Chat = make([]models.Chat, 0, 10)
	for rows.Next() {
		var chat_t models.Chat
		var time_t time.Time
		rows.Scan(&chat_t.ID, &chat_t.Title, &time_t)
		temp := strings.Split(chat_t.Title, "|")
		if temp[0] == user.Login {
			chat_t.Title = temp[1]
		} else {
			chat_t.Title = temp[0]
		}
		chats = append(chats, chat_t)
	}
	return chats, nil
}

func (r *SQLiteRepo) CreateNewChat(ctx context.Context, user1, user2 *models.User) *models.Chat {
	var chat_res models.Chat
	row := r.DB.QueryRowContext(ctx, `SELECT MAX(chat_id) FROM chat_members`)
	row.Scan(&chat_res.ID)
	chat_res.ID++
	chat_res.Title = user1.Login + "|" + user2.Login
	r.DB.ExecContext(ctx, `INSERT INTO chat_members (chat_id, user_id) VALUES (?,?)`, chat_res.ID, user1.Login)
	r.DB.ExecContext(ctx, `INSERT INTO chat_members (chat_id, user_id) VALUES (?,?)`, chat_res.ID, user2.Login)
	r.DB.ExecContext(ctx,
		`INSERT INTO chats
	(id, type, title, created_at)
	VALUES (?, "talk", ?, CURRENT_TIMESTAMP)`,
		chat_res.ID, chat_res.Title)
	return &chat_res
}

func (r *SQLiteRepo) FindOrCreateNewChat(ctx context.Context, user1, user2 *models.User) (*models.Chat, error) {
	_, err := r.GetUserByLogin(ctx, user2.Login)
	if err != nil {
		return nil, err
	}
	res := r.DB.QueryRowContext(ctx,
		`SELECT t1.chat_id
	FROM chat_members t1
	JOIN chat_members t2 ON t1.chat_id=t2.chat_id
	WHERE t1.user_id=? AND t2.user_id=?`, user1.Login, user2.Login)
	chat_res := &models.Chat{}
	if err := res.Scan(&chat_res.ID); err != nil {
		chat_res = r.CreateNewChat(ctx, user1, user2)
	}
	return chat_res, nil
}

func (r *SQLiteRepo) CanUserAccessChat(ctx context.Context, user *models.User, chatID int64) bool {
	res := r.DB.QueryRowContext(ctx,
		`SELECT au.login
		FROM users_auth au
		JOIN chat_members cm ON au.login=cm.user_id
		WHERE au.token=? AND cm.chat_id=?`, user.Token, chatID)
	err := res.Scan(&user.Login)
	return err == nil
}

func (r *SQLiteRepo) SendMessage(ctx context.Context, msg *models.MessageCreateRequest) error {
	_, err := r.DB.Exec(
		`INSERT INTO messages
		(chat_id, sender_id, created_at, body)
		VALUES (?,?,CURRENT_TIMESTAMP,?)`,
		msg.ChatID, msg.Sender, msg.Body)
	res := r.DB.QueryRowContext(ctx, `SELECT MAX(id) FROM messages WHERE chat_id=?`, msg.ChatID)
	res.Scan(&msg.ID)
	return err
}
