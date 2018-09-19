package data

type IBlockService interface {
	CalculateHash(*Block) string
	PrintImpl()
	Print(*Block)
}
