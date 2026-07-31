package domain

// Mensagens de erro e validação do domínio em português.
// Centralizadas aqui para consistência em todas as camadas.

const (
	MsgNameRequired          = "nome é obrigatório"
	MsgCriticalityOutOfRange = "criticalityLevel deve estar entre 1 e 5, recebeu %d"
	MsgLeadTimeNegative      = "leadTimeDays deve ser >= 0, recebeu %d"
	MsgAvgDailySalesNegative = "averageDailySales deve ser >= 0, recebeu %f"
	MsgCriticalityInvalid    = "criticalityLevel deve estar entre 1 e 5"
	MsgLeadTimeInvalid       = "leadTimeDays deve ser >= 0"
	MsgAvgDailySalesInvalid  = "averageDailySales deve ser >= 0"

	MsgPartExists   = "peça com id %s já existe"
	MsgPartNotFound = "peça com id %s não encontrada"

	MsgInvalidBody = "corpo da requisição inválido"
	MsgIDRequired  = "id é obrigatório"
)
