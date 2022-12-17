module traitor/db

go 1.19

require github.com/shopspring/decimal v1.3.1

require (
	traitor/logger v0.0.0
)
replace (
	traitor/logger => ../logger
)