package dotenv

var (
	Host      string
	Name      string
	User      string
	Password  string
	Port      string
	Sslmode   string
	JwtSecret string
)

func Catch() {
	load()

	Host = get("DB_HOST")
	Name = get("DB_NAME")
	User = get("DB_USER")
	Password = get("DB_PASSWORD")
	Port = get("DB_PORT")
	Sslmode = get("DB_SSLMODE")
}
