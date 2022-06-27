containers: {
	"acorn-dns": {
		build: "."
		cmd: "api-server"
		env: {
			SQL_DIALECT: "mysql"
      SQL_DSN: "root:root@tcp(sql:3306)/acorn_dns?charset=utf8mb4&parseTime=True&loc=Local"
      LOGLEVEL: "trace"
		}
		publish: [
			"4315/http",
		],
	},
	sql: {
		image: "mariadb:10.7.4"
		env: {
			MYSQL_ROOT_PASSWORD: "root"
      MYSQL_DATABASE: "acorn_dns"
		}
		ports: [
			"3306/tcp",
		]
	}
}