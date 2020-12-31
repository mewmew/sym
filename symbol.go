package sym

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lunixbochs/struc"
	"github.com/pkg/errors"
)

// A Symbol is a PS1 symbol.
type Symbol struct {
	// Symbol header.
	Hdr *SymbolHeader
	// Symbol body.
	Body SymbolBody
}

// String returns the string representation of the symbol.
func (sym *Symbol) String() string {
	return fmt.Sprintf("%v %v", sym.Hdr, sym.Body)
}

// Size returns the size of the symbol in bytes.
func (sym *Symbol) Size() int {
	hdrSize := binary.Size(*sym.Hdr)
	return hdrSize + sym.Body.BodySize()
}

// A SymbolHeader is a PS1 symbol header.
type SymbolHeader struct {
	// Address or value of symbol.
	Value uint32 `struc:"uint32,little"`
	// Symbol kind; specifies type of symbol body.
	Kind Kind `struc:"uint8"`
}

// String returns the string representation of the symbol header.
func (hdr *SymbolHeader) String() string {
	return fmt.Sprintf("$%08x %v", hdr.Value, hdr.Kind)
}

// SymbolBody is the sum-type of all symbol bodies.
type SymbolBody interface {
	fmt.Stringer
	// BodySize returns the size of the symbol body in bytes.
	BodySize() int
}

// parseSymbol parses and returns a PS1 symbol.
func parseSymbol(r io.Reader) (*Symbol, error) {
	// Parse symbol header.
	sym := &Symbol{}
	hdr, err := parseSymbolHeader(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sym.Hdr = hdr

	// Parse symbol body.
	body, err := parseSymbolBody(r, hdr.Kind)
	if err != nil {
		return sym, errors.WithStack(err)
	}
	sym.Body = body
	return sym, nil
}

// parseSymbolHeader parses and returns a PS1 symbol header.
func parseSymbolHeader(r io.Reader) (*SymbolHeader, error) {
	hdr := &SymbolHeader{}
	if err := struc.Unpack(r, hdr); err != nil {
		return nil, errors.WithStack(err)
	}
	return hdr, nil
}

// parseSymbolBody parses and returns a PS1 symbol body.
func parseSymbolBody(r io.Reader, kind Kind) (SymbolBody, error) {
	parse := func(body SymbolBody) (SymbolBody, error) {
		if err := struc.Unpack(r, body); err != nil {
			return nil, errors.WithStack(err)
		}
		return body, nil
	}
	switch kind {
	case KindName1:
		return parse(&Name1{})
	case KindName2:
		return parse(&Name2{})
	case KindName5:
		return parse(&Name2{})
	case KindName6:
		return parse(&Name2{})
	case KindIncSLD:
		// empty body.
		return &IncSLD{}, nil
	case KindIncSLDByte:
		return parse(&IncSLDByte{})
	case KindIncSLDWord:
		return parse(&IncSLDWord{})
	case KindSetSLD:
		return parse(&SetSLD{})
	case KindSetSLD2:
		return parse(&SetSLD2{})
	case KindEndSLD:
		// empty body.
		return &EndSLD{}, nil
	case KindFuncStart:
		return parse(&FuncStart{})
	case KindFuncEnd:
		return parse(&FuncEnd{})
	case KindBlockStart:
		return parse(&BlockStart{})
	case KindBlockEnd:
		return parse(&BlockEnd{})
	case KindDef:
		return parse(&Def{})
	case KindDef2:
		return parse(&Def2{})
	case KindOverlay:
		return parse(&Overlay{})
	case KindSetOverlay:
		// empty body.
		return &SetOverlay{}, nil
	default:
		return nil, errors.Errorf("support for symbol kind 0x%02X not yet implemented", uint8(kind))
	}
}

// --- [ 0x01 ] ----------------------------------------------------------------

// A Name1 symbol specifies the name of a symbol.
//
// Value of the symbol header specifies the associated address.
type Name1 struct {
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Symbol name,
	Name string
}

// String returns the string representation of the name symbol.
func (body *Name1) String() string {
	// $00000000 1 __RHS2_data_size
	return body.Name
}

// BodySize returns the size of the symbol body in bytes.
func (body *Name1) BodySize() int {
	return 1 + int(body.NameLen)
}

// --- [ 0x02 ] ----------------------------------------------------------------

// A Name2 symbol specifies the name of a symbol.
//
// Value of the symbol header specifies the associated address.
type Name2 struct {
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Symbol name,
	Name string
}

// String returns the string representation of the name symbol.
func (body *Name2) String() string {
	// $80010000 2 printattribute
	return body.Name
}

// BodySize returns the size of the symbol body in bytes.
func (body *Name2) BodySize() int {
	return 1 + int(body.NameLen)
}

// --- [ 0x05 ] ----------------------------------------------------------------

// A Name5 symbol specifies the name of a symbol.
//
// Value of the symbol header specifies the associated address.
type Name5 struct {
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Symbol name,
	Name string
}

// String returns the string representation of the name symbol.
func (body *Name5) String() string {
	// $00000000 5 m
	return body.Name
}

// BodySize returns the size of the symbol body in bytes.
func (body *Name5) BodySize() int {
	return 1 + int(body.NameLen)
}

// --- [ 0x06 ] ----------------------------------------------------------------

// A Name6 symbol specifies the name of a symbol.
//
// Value of the symbol header specifies the associated address.
type Name6 struct {
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Symbol name,
	Name string
}

// String returns the string representation of the name symbol.
func (body *Name6) String() string {
	// $00010604 6 DoTitle
	return body.Name
}

// BodySize returns the size of the symbol body in bytes.
func (body *Name6) BodySize() int {
	return 1 + int(body.NameLen)
}

// --- [ 0x80 ] ----------------------------------------------------------------

// An IncSLD symbol increments the current line number.
//
// Value of the symbol header specifies the associated address.
type IncSLD struct {
}

// String returns the string representation of the line number increment symbol.
func (body *IncSLD) String() string {
	// $80010004 80 Inc SLD linenum (to 116)
	return "Inc SLD linenum"
}

// BodySize returns the size of the symbol body in bytes.
func (body *IncSLD) BodySize() int {
	return 0
}

// --- [ 0x82 ] ----------------------------------------------------------------

// An IncSLDByte symbol specifies the increment of the current line number.
//
// Value of the symbol header specifies the associated address.
type IncSLDByte struct {
	Inc uint8 `struc:"uint8"`
}

// String returns the string representation of the line number increment symbol.
func (body *IncSLDByte) String() string {
	// $80010008 82 Inc SLD linenum by byte 2 (to 118)
	return fmt.Sprintf("Inc SLD linenum by byte %d", body.Inc)
}

// BodySize returns the size of the symbol body in bytes.
func (body *IncSLDByte) BodySize() int {
	return 1
}

// --- [ 0x84 ] ----------------------------------------------------------------

// An IncSLDWord symbol specifies the increment of the current line number.
//
// Value of the symbol header specifies the associated address.
type IncSLDWord struct {
	Inc uint16 `struc:"uint16,little"`
}

// String returns the string representation of the line number increment symbol.
func (body *IncSLDWord) String() string {
	// $80025d38 84 Inc SLD linenum by word 276 (to 479)
	return fmt.Sprintf("Inc SLD linenum by word %d", body.Inc)
}

// BodySize returns the size of the symbol body in bytes.
func (body *IncSLDWord) BodySize() int {
	return 2
}

// --- [ 0x86 ] ----------------------------------------------------------------

// A SetSLD symbol specifies the current line number.
//
// Value of the symbol header specifies the associated address.
type SetSLD struct {
	// Line number.
	Line uint32 `struc:"uint32,little"`
}

// String returns the string representation of the set line number symbol.
func (body *SetSLD) String() string {
	// $8001ff08 86 Set SLD linenum to 88
	return fmt.Sprintf("Set SLD linenum to %d", body.Line)
}

// BodySize returns the size of the symbol body in bytes.
func (body *SetSLD) BodySize() int {
	return 4
}

// --- [ 0x88 ] ----------------------------------------------------------------

// A SetSLD2 symbol specifies the current line number and source file.
//
// Value of the symbol header specifies the associated address.
type SetSLD2 struct {
	// Line number.
	Line uint32 `struc:"uint32,little"`
	// Path length.
	PathLen uint8 `struc:"uint8,sizeof=Path"`
	// Source file,
	Path string
}

// String returns the string representation of the set line number symbol.
func (body *SetSLD2) String() string {
	// $80010000 88 Set SLD to line 115 of file D:\LIB\PSX\NULLFUNC.ASM
	return fmt.Sprintf("Set SLD to line %d of file %s", body.Line, body.Path)
}

// BodySize returns the size of the symbol body in bytes.
func (body *SetSLD2) BodySize() int {
	return 4 + 1 + int(body.PathLen)
}

// --- [ 0x8A ] ----------------------------------------------------------------

// An EndSLD symbol indicates the end of a line number specifier.
//
// Value of the symbol header specifies the associated address.
type EndSLD struct {
}

// String returns the string representation of the end of line number symbol.
func (body *EndSLD) String() string {
	// $80020ffc 8a End SLD info
	return "End SLD info"
}

// BodySize returns the size of the symbol body in bytes.
func (body *EndSLD) BodySize() int {
	return 0
}

// --- [ 0x8C ] ----------------------------------------------------------------

// A FuncStart symbol specifies the start of a function.
//
// Value of the symbol header specifies the associated address.
type FuncStart struct {
	// Frame pointer register.
	FP uint16 `struc:"uint16,little"`
	// Function size.
	FSize uint32 `struc:"uint32,little"`
	// Return address register.
	RetReg uint16 `struc:"uint16,little"`
	// Mask.
	Mask uint32 `struc:"uint32,little"`
	// Mask offset.
	MaskOffset int32 `struc:"int32,little"`
	// Line number.
	Line uint32 `struc:"uint32,little"`
	// Path length.
	PathLen uint8 `struc:"uint8,sizeof=Path"`
	// Source file.
	Path string
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Symbol name.
	Name string
}

// String returns the string representation of the function start symbol.
func (body *FuncStart) String() string {
	// $8001fefc 8c Function start
	//    fp = 29
	//    fsize = 24
	//    retreg = 31
	//    mask = $80000000
	//    maskoffs = -8
	//    line = 88
	//    file = C:\DIABPSX\GLIBDEV\SOURCE\TASKER.C
	//    name = DoEpi
	const format = `Function start
    fp = %d
    fsize = %d
    retreg = %d
    mask = $%08x
    maskoffs = %d
    line = %d
    file = %s
    name = %s`
	return fmt.Sprintf(format, body.FP, body.FSize, body.RetReg, body.Mask, body.MaskOffset, body.Line, body.Path, body.Name)
}

// BodySize returns the size of the symbol body in bytes.
func (body *FuncStart) BodySize() int {
	return 2 + 4 + 2 + 4 + 4 + 4 + 1 + int(body.PathLen) + 1 + int(body.NameLen)
}

// --- [ 0x8E ] ----------------------------------------------------------------

// A FuncEnd symbol specifies the end of a function.
//
// Value of the symbol header specifies the associated address.
type FuncEnd struct {
	// Line number.
	Line uint32 `struc:"uint32,little"`
}

// String returns the string representation of the function end symbol.
func (body *FuncEnd) String() string {
	// $8001ff4c 8e Function end   line 91
	return fmt.Sprintf("Function end   line %d", body.Line)
}

// BodySize returns the size of the symbol body in bytes.
func (body *FuncEnd) BodySize() int {
	return 4
}

// --- [ 0x90 ] ----------------------------------------------------------------

// A BlockStart symbol specifies the start of a block.
//
// Value of the symbol header specifies the associated address.
type BlockStart struct {
	// Line number.
	Line uint32 `struc:"uint32,little"`
}

// String returns the string representation of the block start symbol.
func (body *BlockStart) String() string {
	// $8003017c 90 Block start  line = 1
	return fmt.Sprintf("Block start  line = %d", body.Line)
}

// BodySize returns the size of the symbol body in bytes.
func (body *BlockStart) BodySize() int {
	return 4
}

// --- [ 0x92 ] ----------------------------------------------------------------

// A BlockEnd symbol specifies the end of a block.
//
// Value of the symbol header specifies the associated address.
type BlockEnd struct {
	// Line number.
	Line uint32 `struc:"uint32,little"`
}

// String returns the string representation of the block end symbol.
func (body *BlockEnd) String() string {
	// $800301ac 92 Block end  line = 4
	return fmt.Sprintf("Block end  line = %d", body.Line)
}

// BodySize returns the size of the symbol body in bytes.
func (body *BlockEnd) BodySize() int {
	return 4
}

// --- [ 0x94 ] ----------------------------------------------------------------

// A Def symbol specifies the class, type, size and name of a definition.
//
// Value of the symbol header specifies the associated address.
type Def struct {
	// Definition class.
	Class Class `struc:"uint16,little"`
	// Definition type.
	Type Type `struc:"uint16,little"`
	// Definition size.
	Size uint32 `struc:"uint32,little"`
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Definition name,
	Name string
}

// String returns the string representation of the definition symbol.
func (body *Def) String() string {
	// $00000000 94 Def class TPDEF type UCHAR size 0 name u_char
	return fmt.Sprintf("Def class %v type %v size %v name %v", body.Class, body.Type, body.Size, body.Name)
}

// BodySize returns the size of the symbol body in bytes.
func (body *Def) BodySize() int {
	return 2 + 2 + 4 + 1 + int(body.NameLen)
}

// --- [ 0x96 ] ----------------------------------------------------------------

// A Def2 symbol specifies the class, type, size, dimensions, tag and name of a
// definition.
//
// Value of the symbol header specifies the associated address.
type Def2 struct {
	// Definition class.
	Class Class `struc:"uint16,little"`
	// Definition type.
	Type Type `struc:"uint16,little"`
	// Definition size.
	Size uint32 `struc:"uint32,little"`
	// Dimensions length.
	DimsLen uint16 `struc:"uint16,little,sizeof=Dims"`
	// Dimensions.
	Dims []uint32 `struc:"[]uint32,little"`
	// Tag length.
	TagLen uint8 `struc:"uint8,sizeof=Tag"`
	// Definition tag,
	Tag string
	// Name length.
	NameLen uint8 `struc:"uint8,sizeof=Name"`
	// Definition name,
	Name string
}

// String returns the string representation of the definition symbol.
func (body *Def2) String() string {
	// $00000000 96 Def2 class MOS type ARY INT size 4 dims 1 1 tag  name r
	dd := make([]string, len(body.Dims))
	for i, dim := range body.Dims {
		dd[i] = strconv.Itoa(int(dim))
	}
	dims := fmt.Sprintf("%d %s", body.DimsLen, strings.Join(dd, " "))
	if body.DimsLen == 0 {
		dims = "0"
	}
	return fmt.Sprintf("Def2 class %v type %v size %v dims %s tag %v name %v", body.Class, body.Type, body.Size, dims, body.Tag, body.Name)
}

// BodySize returns the size of the symbol body in bytes.
func (body *Def2) BodySize() int {
	return 2 + 2 + 4 + 2 + int(4*body.DimsLen) + 1 + int(body.TagLen) + 1 + int(body.NameLen)
}

// --- [ 0x98 ] ----------------------------------------------------------------

// An Overlay symbol specifies the length and id of a file overlay (e.g. a
// shared library).
//
// Value of the symbol header specifies the base address at which the overlay is
// loaded.
type Overlay struct {
	// Overlay length in bytes.
	Length uint32 `struc:"uint32,little"`
	// Overlay ID.
	ID uint32 `struc:"uint32,little"`
}

// String returns the string representation of the overlay symbol.
func (body *Overlay) String() string {
	// $800b031c overlay length $000009e4 id $4
	return fmt.Sprintf("length $%08x id $%x", body.Length, body.ID)
}

// BodySize returns the size of the symbol body in bytes.
func (body *Overlay) BodySize() int {
	return 4 + 4
}

// --- [ 0x9A ] ----------------------------------------------------------------

// A SetOverlay specifies the active overlay.
//
// Value of the symbol header specifies the active overlay ID.
type SetOverlay struct {
}

// String returns the string representation of the set overlay symbol.
func (body *SetOverlay) String() string {
	// $00000004 set overlay
	return ""
}

// BodySize returns the size of the symbol body in bytes.
func (body *SetOverlay) BodySize() int {
	return 0
}
