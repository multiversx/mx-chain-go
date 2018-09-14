package data

type IBlockService interface {
	CalculateHash(block Block) string
	PrintImpl()
}
