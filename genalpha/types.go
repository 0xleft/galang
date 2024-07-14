package genalphatypes

type ASTNodeType int

const (
	ASTNodeTypeProgram ASTNodeType = iota
	ASTNodeTypeExpression
	ASTNodeTypeFunctionDeclaration
	ASTNodeTypeFunctionCall
	ASTNodeTypeVariableDeclaration
	ASTNodeTypeVariableAssignment
	ASTNodeTypeIf
	ASTNodeTypeImport
	ASTNodeTypeWhile
	ASTNodeTypeReturn
	ASTNodeTypeOperator
	ASTNodeTypeBinaryOperation
	ASTNodeTypeUnaryOperation
	ASTNodeTypeIdentifier
	ASTNodeTypeNumber
	ASTNodeTypeString
	ASTNodeTypeBoolean
	ASTNodeTypeNone
	ASTNodeTypeFunctionArgument
	ASTNodeTypeMemberAssignment
	ASTNodeTypeMemberAccess
	ASTNodeTypeBlock
	ASTNodeTypeUnknown
)

type ASTNode struct {
	Type     ASTNodeType
	Children []ASTNode
	Value    string
}

type TokenType int

const (
	TokenTypeIdentifier TokenType = iota
	TokenTypeNumber
	TokenTypeString
	TokenTypeBoolean
	TokenTypeKeyword
	TokenTypeOperator
	TokenTypePunctuation
	TokenTypeComment
	TokenTypeWhitespace
	TokenTypeNewline
	TokenTypeUnknown
)

type Keyword string

const (
	KeywordVar    Keyword = "fax"
	KeywordIf     Keyword = "foreal"
	KeywordIfYes  Keyword = "yeah"
	KeywordIfNo   Keyword = "nah"
	KeywordFunc   Keyword = "lowkey"
	KeywordEnd    Keyword = "end"
	KeywordCall   Keyword = "fire"
	KeywordWhile  Keyword = "durin"
	KeywordImport Keyword = "gyat"
	KeywordReturn Keyword = "rizzult"
	KeywordTrue   Keyword = "yay"
	KeywordFalse  Keyword = "nay"
	KeywordNone   Keyword = "nuthin"
)

var (
	Keywords = []string{
		string(KeywordVar),
		string(KeywordIf),
		string(KeywordIfYes),
		string(KeywordIfNo),
		string(KeywordFunc),
		string(KeywordEnd),
		string(KeywordCall),
		string(KeywordWhile),
		string(KeywordImport),
		string(KeywordReturn),
		string(KeywordTrue),
		string(KeywordFalse),
		string(KeywordNone),
	}
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
}
