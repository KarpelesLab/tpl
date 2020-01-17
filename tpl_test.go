package tpl_test

import (
	"context"
	"os"
	"testing"

	"github.com/KarpelesLab/tpl"
)

var tplTestVars = []struct {
	in  string
	out string
}{
	// tests from the PHP version
	{`Hello world`, `Hello world`},
	{`{{if "1" == "0"}}This will not appear{{elseif {{TEST}}=="World"}}This will appear{{else}}With won't{{/if}}`, `This will appear`},
	{`{{if {{TEST}} == "World"}}This OK{{else}}BAD{{/if}}`, `This OK`},
	{`{{if {{_TPL_PAGE}}}}ok{{else}}bad{{/if}}`, `ok`},
	{`{{set _TEST="JP"}}{{if {{_TEST}}}}OK: {{_TEST}}{{else}}BAD: {{_TEST}}{{/if}}{{/set}}`, `OK: JP`},
	{`This is page {{_TPL_PAGE}}`, `This is page index.html`},
	{`Printf {{@printf("this is %s","SPARTA")}}`, `Printf this is SPARTA`},
	{`Test: {{foreach {{@seq("1","3")}} as _X}}This is {{_X}} ha{{/foreach}}`, `Test: This is 1 haThis is 2 haThis is 3 ha`},
	//'Test API calls: {{@rest("Misc/Debug:fixedString")}}' => 'Test API calls: fixed string',
	{`a{{literal}}{{if "1" == "0"}}This is a test{{/if}}{{/literal}}b`, `a{{if "1" == "0"}}This is a test{{/if}}b`},
	{`{{@string("{{literal}}"{{/literal}}")}}`, `"`},
	//{`{{try}}Hello {{@rest("Misc/Debug:error")}}{{catch}}Got error {{_EXCEPTION}}{{/try}}`,`Got error Test error`},
	//{`{{foreach {{@string("A")|unicode()}} as _C}}{{_C/Name}}{{/foreach}}`, `LATIN CAPITAL LETTER A`},
	{`{{set _X="hello world"}}Test {{_X}}!! {{_X}}!!!{{/set}}`, `Test hello world!! hello world!!!`},
	{`{{set _X="hello world" _Y="REALLY!"}}Test {{_X}}!! {{_Y}}!{{/set}}`, `Test hello world!! REALLY!!`},
	{`{{set _X={{@string("a,b,c,d,e")|explode(",")}}}}{{foreach {{_X}} as _Y}}{{_Y}}{{/foreach}}{{/set}}`, `abcde`},
	//{`{{set _X={{@string("a,b,a,b,c")|explode(",")}}}}{{foreach {{_X|arrayfilter("","a")}} as _Y}}{{_Y}}{{/foreach}}{{/set}}`, `aa`},
	//{`{{set _X={{@string("a,,a,,c")|explode(",")}}}}{{foreach {{_X|arrayfilter("","false")}} as _Y}}{{_Y}} {{/foreach}}{{/set}}`, `  `},
	{`{{set _X={{@string("{{literal}}{"a":"b"}{{/literal}}")|jsonparse()}}}}test: {{foreach {{_X}} as _Y}}{{_Y_KEY}}={{_Y}}{{/foreach}}.{{/set}}`, `test: a=b.`},
	{`Test: {{set _X=(1 + 2)}}value: {{_X}}{{/set}}`, `Test: value: 3`},
	{`Test2: {{set _X=(1 && 2)}}value: {{_X|dump()}}{{/set}}`, `Test2: value: interfaceValue[bool(true)]`},
	{`Test3: {{set _X=({{_TEST_ARRAY}} && {{_TPL_PAGE}})}}value: {{_X|dump()}}{{/set}}`, `Test3: value: interfaceValue[bool(true)]`},
	{" \n{{@string(\"a\")}}\n\n{{@string(\"b\")}}\n", " \na\n\nb\n"},

	// our tests
	{"<b>Hello {{TEST}}!</b>", "<b>Hello World!</b>"},
	{"Some language: {{@string(\"日本語 \\\" json</d>\")|json()}}", `Some language: "日本語 \" json\u003c/d\u003e"`},
	{`If test {{if 0}}ignored part{{else}}re-{{/if}}run`, `If test re-run`},
	{`entities filter: {{@string("<b>entities'here\"</b>")|entities()}}`, `entities filter: &lt;b&gt;entities&#39;here&#34;&lt;/b&gt;`},
	{"Complex filters: {{TEST|entities()|explode(\"o\")|implode(\"O\")|json()}}", `Complex filters: "WOrld"`},
	{`Substr: {{@string("Hello world")|substr(2, 8)}}`, `Substr: llo worl`},
	{`Substr: {{@string("Hello world")|substr("-5","-1")}}`, `Substr: worl`},
	{`Maths: {{@string(1 + 2)}}`, `Maths: 3`},
	{`Priorities: {{@string(1 + 2 * 3)}}`, `Priorities: 7`},
	{`Priorities2: {{@string((1 + 2) * 3)}}`, `Priorities2: 9`},
	{`Floats: {{@string(1 + 2.7)}}`, `Floats: 3.7`},
	{`Bool: {{@string(!0)}}`, `Bool: 1`},
	{`Json: {{@string("\"hello\"")|jsonparse()}}`, `Json: hello`},
	{`Seq: {{foreach {{@seq(1, 3)}} as _X}}{{_X}}.{{/foreach}}`, `Seq: 1.2.3.`},
	{`Empty: {{foreach {{@seq(1, 0)}} as _X}}bad{{else}}good{{/foreach}}`, `Empty: good`},
	{`Var set: {{@string("{{_X}}")|_X="test"}}`, `Var set: test`},
	{`foreach literal list: {{foreach (1,2,3,6*7,"toto",{{TEST}}) as _X}}{{_X}}-{{/foreach}}`, `foreach literal list: 1-2-3-42-toto-World-`},
	{`Exceptions: {{try}}try: {{@error("test %s", "yes")}}{{catch _E}}catch: {{_E}}{{/try}}`, `Exceptions: catch: At main on line 1 (position 25): function call failed: test yes`},
	{`if test: {{if !{{_TPL_PAGE}}}}bug: {{_TPL_PAGE}}{{/if}}`, `if test: `},
	{`Test: {{GET|_VALUE="foo"}}`, `Test: Value:foo`},
	{`Loop {{foreach {{_TEST_CMPLX}} as _X}}{{if {{_X/a}}=="foo"}}{{_X/a}}{{/if}}{{/foreach}}`, `Loop foo`},

	// array tests
	{`Array to Json: {{_TEST_ARRAY|json()}}`, `Array to Json: ["hello","world"]`},
	{`Array offset: {{_TEST_ARRAY/0|json()}}`, `Array offset: "hello"`},
	{`Array offset 1: {{_TEST_ARRAY/1|json()}}`, `Array offset 1: "world"`},
	{`Obj to json: {{_TEST_IDX|json()}}`, `Obj to json: {"Foo":"bar"}`},
	{`Obj offset: {{_TEST_IDX/Foo}}`, `Obj offset: bar`},

	// filter tests
	{`Array Reverse: {{_TEST_ARRAY|reverse()|json()}}`, `Array Reverse: ["world","hello"]`},
	{`String Reverse: {{_TPL_PAGE|reverse()}}`, `String Reverse: lmth.xedni`},
	{`Trunc: {{@string("hello world this is some thing")|truncate(20)}}`, `Trunc: hello world this is…`},
	{`Round: {{_TEST_FLOAT|round()}}`, `Round: 3.14`},
	{`Round: {{_TEST_FLOAT|round(1)}}`, `Round: 3.1`},
	{`Round: {{_TEST_FLOAT|round(3)}}`, `Round: 3.142`},
	{`Round: {{_TEST_FLOAT|round(0)}}`, `Round: 3`},
	{`Array slice: {{_TEST_ARRAY|arrayslice(0,1)|json()}}`, `Array slice: ["hello"]`},
	{`Array slice: {{_TPL_PAGE|arrayslice(5)}}`, `Array slice: index`},
	{`Array Filter: {{_TEST_CMPLX|arrayfilter("a","foo")|json()}}`, `Array Filter: [{"a":"foo"}]`},
	{`toInt: {{_TEST_FLOAT|toint()|dump()}}`, `toInt: interfaceValue[int64(3)]`},
	{`StripTag: {{@string("<b>this is bold!</b>")|striptags()}}`, `StripTag: this is bold!`},
	{`dur1: {{@string("86400")|duration()}}`, `dur1: 1:00:00:00`},
	{`dur2: {{@string("86399")|duration()}}`, `dur2: 23:59:59`},
	{`dur3: {{@string("12345678")|duration()}}`, `dur3: 142:21:21:18`},
	{`s: {{foreach {{@seq(1,9)|lines(3)}} as _A}}[{{foreach {{_A}} as _B}}={{_B}}{{/foreach}}]{{/foreach}}!`, `s: [=1=2=3][=4=5=6][=7=8=9]!`},
	{`s: {{foreach {{@seq(1,9)|columns(3)}} as _A}}[{{foreach {{_A}} as _B}}={{_B}}{{/foreach}}]{{/foreach}}!`, `s: [=1=4=7][=2=5=8][=3=6=9]!`},
}

func TestTpl(t *testing.T) {
	e := tpl.New()
	e.Raw.TemplateData["test"] = "World"
	e.Raw.TemplateData["get"] = "Value:{{_VALUE}}"

	ctx := context.Background()
	ctx = tpl.ValuesCtx(ctx, map[string]interface{}{
		"_tpl_page":   "index.html",
		"_test_float": 3.14159265359,
		"_test_array": tpl.Values{tpl.NewValue("hello"), tpl.NewValue("world")},
		"_test_cmplx": tpl.Values{tpl.NewValue(map[string]tpl.Value{"a": tpl.NewValue("foo")}), tpl.NewValue(map[string]tpl.Value{"a": tpl.NewValue("bar")})},
		"_test_idx":   map[string]tpl.Value{"Foo": tpl.NewValue("bar")},
	})

	for _, x := range tplTestVars {
		e.Raw.TemplateData["main"] = x.in
		cErr := e.Compile(ctx)

		if cErr != nil {
			t.Errorf("Test compile failed: %s", cErr)
			continue
		}

		if res, err := e.ParseAndReturn(ctx, "main"); res != x.out {
			e.Dump(os.Stdout, 0)
			t.Errorf("Test %#v should equal %#v (got: %#v:%v)", x.in, x.out, res, err)
		}
	}
}
