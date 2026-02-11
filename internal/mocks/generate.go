package mocks

//go:generate go run go.uber.org/mock/mockgen@latest -destination=order_repository.go -package=mocks loyalty/internal/domain/order/repository OrdersRepository
//go:generate go run go.uber.org/mock/mockgen@latest -destination=order_service.go -package=mocks loyalty/internal/domain/order/service OrdersService,OrderNumberValidator
//go:generate go run go.uber.org/mock/mockgen@latest -destination=accrual_client.go -package=mocks loyalty/internal/domain/accrual/client AccrualClient
//go:generate go run go.uber.org/mock/mockgen@latest -destination=auth_user_repository.go -package=mocks loyalty/internal/domain/auth/repository UserRepository
//go:generate go run go.uber.org/mock/mockgen@latest -destination=auth_services.go -package=mocks loyalty/internal/domain/auth/service AuthService,UserService,TokenService
//go:generate go run go.uber.org/mock/mockgen@latest -destination=auth_usecase.go -package=mocks loyalty/internal/domain/auth/usecase AuthUsecase
//go:generate go run go.uber.org/mock/mockgen@latest -destination=balance_repository.go -package=mocks loyalty/internal/domain/balance/repository BalanceRepository
//go:generate go run go.uber.org/mock/mockgen@latest -destination=balance_service.go -package=mocks loyalty/internal/domain/balance/service BalanceService
//go:generate go run go.uber.org/mock/mockgen@latest -destination=balance_usecase.go -package=mocks loyalty/internal/domain/balance/usecase BalanceUsecase
//go:generate go run go.uber.org/mock/mockgen@latest -destination=order_usecase.go -package=mocks loyalty/internal/domain/order/usecase OrdersUsecase
//go:generate go run go.uber.org/mock/mockgen@latest -destination=withdrawal_repository.go -package=mocks loyalty/internal/domain/withdrawal/repository AccountRepository,WithdrawalsRepository
//go:generate go run go.uber.org/mock/mockgen@latest -destination=withdrawal_service.go -package=mocks loyalty/internal/domain/withdrawal/service WithdrawalsService
//go:generate go run go.uber.org/mock/mockgen@latest -destination=withdrawal_usecase.go -package=mocks loyalty/internal/domain/withdrawal/usecase WithdrawalsUsecase
