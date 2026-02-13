package route

type OrderField string

const (
	OrderFieldCreatedAt OrderField = "created_at"
	OrderFieldPath      OrderField = "path"
	OrderFieldTargetURL OrderField = "target_url"
)
