# atlas.hcl
# 定义外部 schema 数据源
data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./models",
    "--dialect", "postgres",
  ]
}

# 开发环境（用于生成迁移文件）
env "dev" {
  src = data.external_schema.gorm.url
  url = "postgres://postgres:postgres@localhost:5432/xcoding_code_repository?sslmode=disable&search_path=public"
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
}


# 生产环境（仅用于应用迁移）
env "prod" {
  # 生产数据库连接（这是真正要操作的数据厍）
  url = "postgres://postgres:postgres@localhost:5432/xcoding_code_repository?sslmode=disable&search_path=public"

  # 迁移文件目录（从版本控制获取）
  migration {
    dir = "file://migrations"
  }
}