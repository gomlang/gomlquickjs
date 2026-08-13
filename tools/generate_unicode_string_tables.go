package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type tableSpec struct {
	source string
	field  string
	type_  string
}

func main() {
	if len(os.Args) != 3 {
		panic("expected input and output paths")
	}
	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	specs := []tableSpec{
		{"case_conv_table1", "case_table1", "uint32"},
		{"case_conv_table2", "case_table2", "uint8"},
		{"case_conv_ext", "case_extension", "uint16"},
		{"unicode_prop_Cased1_table", "cased_table", "uint8"},
		{"unicode_prop_Cased1_index", "cased_index", "uint8"},
		{"unicode_prop_Case_Ignorable_table", "case_ignorable_table", "uint8"},
		{"unicode_prop_Case_Ignorable_index", "case_ignorable_index", "uint8"},
		{"unicode_cc_table", "combining_class_table", "uint8"},
		{"unicode_cc_index", "combining_class_index", "uint8"},
		{"unicode_decomp_table1", "decomposition_table1", "uint32"},
		{"unicode_decomp_table2", "decomposition_table2", "uint16"},
		{"unicode_decomp_data", "decomposition_data", "uint8"},
		{"unicode_comp_table", "composition_table", "uint16"},
	}
	var output strings.Builder
	output.WriteString("package unicode;\n\n")
	output.WriteString("fn string_hex_nibble(value: byte) -> uint32 {\n")
	output.WriteString("    if value >= b'0' && value <= b'9' {\n")
	output.WriteString("        (value - b'0').to_uint32()\n")
	output.WriteString("    } else {\n")
	output.WriteString("        (value - b'a' + 10).to_uint32()\n")
	output.WriteString("    }\n")
	output.WriteString("}\n\n")
	for _, type_ := range []string{"uint8", "uint16", "uint32"} {
		width := map[string]int{"uint8": 2, "uint16": 4, "uint32": 8}[type_]
		fmt.Fprintf(&output, "fn string_decode_%s(input: string) -> Vec[%s] {\n", type_, type_)
		fmt.Fprintf(&output, "    let output: Vec[%s] = Vec::with_capacity(input.byte_len() / %d);\n", type_, width)
		output.WriteString("    let mut index = 0;\n")
		output.WriteString("    while index < input.byte_len() {\n")
		output.WriteString("        let mut value: uint32 = 0;\n")
		fmt.Fprintf(&output, "        let end = index + %d;\n", width)
		output.WriteString("        while index < end {\n")
		output.WriteString("            value = (value << 4) | string_hex_nibble(input.byte_get(index));\n")
		output.WriteString("            index += 1;\n")
		output.WriteString("        }\n")
		if type_ == "uint32" {
			output.WriteString("        output.push(value);\n")
		} else {
			fmt.Fprintf(&output, "        output.push(value.to_%s());\n", type_)
		}
		output.WriteString("    }\n")
		output.WriteString("    output\n")
		output.WriteString("}\n\n")
	}
	for _, spec := range specs {
		pattern := regexp.MustCompile(`(?s)static const ` + spec.type_ + `_t ` + spec.source + `\[[^]]+\] = \{(.*?)\};`)
		match := pattern.FindSubmatch(input)
		if match == nil {
			panic("missing table " + spec.source)
		}
		body := regexp.MustCompile(`(?m)//.*$`).ReplaceAll(match[1], nil)
		body = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAll(body, nil)
		values := strings.Split(string(body), ",")
		width := map[string]int{"uint8": 2, "uint16": 4, "uint32": 8}[spec.type_]
		var encoded strings.Builder
		for _, raw := range values {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			parsed, err := strconv.ParseUint(value, 0, 32)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(&encoded, "%0*x", width, parsed)
		}
		fmt.Fprintf(&output, "fn string_%s() -> Vec[%s] {\n", spec.field, spec.type_)
		fmt.Fprintf(&output, "    string_decode_%s(\"%s\")\n", spec.type_, encoded.String())
		output.WriteString("}\n\n")
	}
	output.WriteString("fn string_unicode_data() -> StringUnicodeData {\n")
	output.WriteString("    StringUnicodeData {\n")
	for _, spec := range specs {
		fmt.Fprintf(&output, "        %s: string_%s(),\n", spec.field, spec.field)
	}
	output.WriteString("    }\n}\n")
	if err := os.WriteFile(os.Args[2], []byte(output.String()), 0644); err != nil {
		panic(err)
	}
}
