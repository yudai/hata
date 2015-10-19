package hata

import (
	"reflect"
	"testing"
)

func TestSplitCamelCase(t *testing.T) {
	checkCamelCaseEqual(t, "XmlReader", []string{"Xml", "Reader"})
	checkCamelCaseEqual(t, "XMLReader", []string{"XML", "Reader"})
	checkCamelCaseEqual(t, "SuperXMLReader", []string{"Super", "XML", "Reader"})
	checkCamelCaseEqual(t, "Http2", []string{"Http2"})
	checkCamelCaseEqual(t, "HTTP2", []string{"HTTP2"})
	checkCamelCaseEqual(t, "Http2Reader", []string{"Http2", "Reader"})
	checkCamelCaseEqual(t, "HTTP2Reader", []string{"HTTP2", "Reader"})
	checkCamelCaseEqual(t, "L2TP", []string{"L2TP"})
	checkCamelCaseEqual(t, "SuperL2TPReader", []string{"Super", "L2TP", "Reader"})
	checkCamelCaseEqual(t, "L2tp", []string{"L2tp"})
	checkCamelCaseEqual(t, "SuperL2tpReader", []string{"Super", "L2tp", "Reader"})
	checkCamelCaseEqual(t, "Layer3xHTTP39Version2", []string{"Layer3x", "HTTP39", "Version2"})
	checkCamelCaseEqual(t, "ReaderX", []string{"Reader", "X"})
	checkCamelCaseEqual(t, "ReaderXML", []string{"Reader", "XML"})
	checkCamelCaseEqual(t, "lowerReader", []string{"lower", "Reader"})
}

func checkCamelCaseEqual(t *testing.T, str string, expected []string) {
	if !reflect.DeepEqual(SplitCamelCase(str), expected) {
		t.Fail()
	}
}

type Formats struct {
	Format       string `short:"f"`
	LongerFormat string

	CombineA bool `short:"a"`
	CombineB bool `short:"b"`
	CombineC bool `short:"c"`

	NotRequireValue bool
}

type DataTypes struct {
	Int   int
	Int8  int8
	Int16 int16
	Int32 int32
	Int64 int64

	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Uintptr uintptr

	Float32 float32
	Float64 float64

	String string

	SliceInt []int
	Map      map[string]int
	Ptr      *string
}

type GlobalOption struct {
	Port string
}

type CmdAOption struct {
	Mode string
}

type CmdBOption struct {
	Type string
}

func TestContext(t *testing.T) {
	g := GlobalOption{}
	c, _ := NewParser(&g, []string{"cmda", "cmdb"})
	c.Parse([]string{})
}

func TestFlagFormats(t *testing.T) {
	v := Formats{}
	c, err := NewParser(&v, []string{})
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	parse(t, c, []string{"--format=foo"})
	if v.Format != "foo" {
		t.Fatal()
	}

	parse(t, c, []string{"--format", "baz"})
	if v.Format != "baz" {
		t.Fatal()
	}

	parse(t, c, []string{"-fbar"})
	if v.Format != "bar" {
		t.Fatal()
	}

	parse(t, c, []string{"-f", "bar"})
	if v.Format != "bar" {
		t.Fatal()
	}

	parse(t, c, []string{"--longer-format", "foo"})
	if v.LongerFormat != "foo" {
		t.Fatal()
	}

	parse(t, c, []string{"--longer-format=bar"})
	if v.LongerFormat != "bar" {
		t.Fatal()
	}

	parse(t, c, []string{"--longer-format=foo", "-f", "bar"})
	if v.LongerFormat != "foo" {
		t.Fatal()
	}
	if v.Format != "bar" {
		t.Fatal()
	}

	// combined
	parse(t, c, []string{"-abc"})
	if v.CombineA != true || v.CombineB != true || v.CombineC != true {
		t.Fatal()
	}

	// errors
	var errs []error
	_, errs = c.Parse([]string{"--format", "--longer-format", "v"})
	if len(errs) != 1 || errs[0].Error() != "Argument `format` requires a value, but not provided" {
		t.Fatal(errs[0])
	}

	_, errs = c.Parse([]string{"--format", "--longer-format"})
	if len(errs) != 2 ||
		errs[0].Error() != "Argument `format` requires a value, but not provided" ||
		errs[1].Error() != "Argument `longer-format` requires a value, but not provided" {
		t.Fatal(errs[0])
	}

	_, errs = c.Parse([]string{"-af"})
	if len(errs) != 1 ||
		errs[0].Error() != "Argument `f` is not a binary flag, but found in a combined flag." {
		t.Fatal(errs[0])
	}

	_, errs = c.Parse([]string{"--format=1", "--format=2"})
	if len(errs) != 1 ||
		errs[0].Error() != "Argument `format` can not accept multiple values" {
		t.Fatal(errs[0])
	}

}

func TestDataTypes(t *testing.T) {
	v := DataTypes{}

	c, err := NewParser(&v, []string{})
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	parse(t, c, []string{"--int=1"})
	if v.Int != 1 {
		t.Fatal("int")
	}
	parse(t, c, []string{"--int8=1"})
	if v.Int8 != 1 {
		t.Fatal("int8")
	}
	parse(t, c, []string{"--int16=1"})
	if v.Int16 != 1 {
		t.Fatal("int16")
	}
	parse(t, c, []string{"--int32=1"})
	if v.Int32 != 1 {
		t.Fatal("int32")
	}
	parse(t, c, []string{"--int64=1"})
	if v.Int64 != 1 {
		t.Fatal("int64")
	}

	parse(t, c, []string{"--uint=1"})
	if v.Uint != 1 {
		t.Fatal("uint")
	}
	parse(t, c, []string{"--uint8=1"})
	if v.Uint8 != 1 {
		t.Fatal("uint8")
	}
	parse(t, c, []string{"--uint16=1"})
	if v.Uint16 != 1 {
		t.Fatal("uint16")
	}
	parse(t, c, []string{"--uint32=1"})
	if v.Uint32 != 1 {
		t.Fatal("uint32")
	}
	parse(t, c, []string{"--uint64=1"})
	if v.Uint64 != 1 {
		t.Fatal("uint64")
	}
	parse(t, c, []string{"--uintptr=1"})
	if v.Uintptr != 1 {
		t.Fatal("uintptr")
	}

	parse(t, c, []string{"--float32=1.0"})
	if v.Float32 != 1.0 {
		t.Fatal("float32")
	}
	parse(t, c, []string{"--float64=1.0"})
	if v.Float64 != 1.0 {
		t.Fatal("float64")
	}

	parse(t, c, []string{"--string=foobar"})
	if v.String != "foobar" {
		t.Fatal("string")
	}

	parse(t, c, []string{"--slice-int=100", "--slice-int=200"})
	if len(v.SliceInt) != 2 || v.SliceInt[0] != 100 || v.SliceInt[1] != 200 {
		t.Fatal("sliceInt")
	}

	parse(t, c, []string{"--map=first:1", "--map=second:2"})
	if len(v.Map) != 2 || v.Map["first"] != 1 || v.Map["second"] != 2 {
		t.Fatal("map")
	}

	parse(t, c, []string{"--ptr=hoge"})
	if *v.Ptr != "hoge" {
		t.Fatal("Ptr")
	}

	// errors
	var errs []error
	_, errs = c.Parse([]string{"--int=something"})
	if len(errs) != 1 || errs[0].Error() != "Failed to parse `int`: Non numeric value is provided to numeric parameter" {
		t.Fatal("int=string")
	}
	_, errs = c.Parse([]string{"--int8=100000000000000"})
	if len(errs) != 1 || errs[0].Error() != "Failed to parse `int8`: Given value is out of range" {
		t.Fatal("int8=overflow")
	}
}

func testDuplicates(t *testing.T) {
	v := &DataTypes{}
	c, err := NewParser(&v, []string{})
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	_, errs := c.Parse([]string{"--int=1", "--int=2"})
	if len(errs) != 1 || errs[0].Error() != "Duplicated arguments found: int" {
		t.Fatal()
	}
}

func testStopWords(t *testing.T) {
	var v *Formats
	var c *Parser
	var err error
	var remaining []string

	v = &Formats{}
	c, err = NewParser(&v, []string{"stop0", "stop1"})
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	remaining, errs := c.Parse([]string{"stop1", "-x", "1", "-y", "2"})
	if !reflect.DeepEqual(remaining, []string{"-x", "1", "-y", "2"}) {
		t.Fatal()
	}
	if len(errs) != 0 {
		t.Fatal()
	}

	v = &Formats{}
	c, err = NewParser(&v, []string{"stop"})
	remaining, errs = c.Parse([]string{"--format", "foo", "stop", "--long-format", "bar"})
	if !reflect.DeepEqual(remaining, []string{"--long-format", "bar"}) {
		t.Fatal()
	}
	if len(errs) != 0 {
		t.Fatal()
	}
	if v.Format != "foo" {
		t.Fatal()
	}
	if v.LongerFormat != "" {
		t.Fatal()
	}
}

func parse(t *testing.T, c *Parser, args []string) {
	if _, errs := c.Parse(args); len(errs) != 0 {
		t.Fatalf("Error: %s", errs[0].Error())
	}
}
