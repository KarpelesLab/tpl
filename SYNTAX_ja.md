# テンプレートエンジン構文ドキュメント

このドキュメントはTPL-GOテンプレートエンジンの構文と機能について説明します。

## 基本的なテンプレート構文

### デリミタ
- テンプレートでは `{{` と `}}` をデリミタとして式や制御構造をマークします
- デリミタの外側の静的なテキストはそのまま出力されます
- バックスラッシュでデリミタをエスケープできます: `\{{` で文字通りの `{{` を出力
- `{{literal}}...{{/literal}}` を使用して囲まれたコンテンツを処理せずにリテラルテキストとして出力します

### テンプレートのインクルード
- アンダースコアのないプレフィックスの `{{NAME}}` は別のテンプレートファイルをインクルードします（例: `{{HEADER}}` は header.tpl をインクルード）
- テンプレートに直接パラメータを渡すことはできません
- インクルードされるテンプレートの変数は、setブロックを使用して設定する必要があります:
  ```
  {{set _X="abc"}}
    {{HEADER}}
  {{/set}}
  ```
- インクルード前に設定された変数は、インクルードされたテンプレート内で利用可能になります

### 変数
- 変数はデリミタ内でアクセスします: `{{_VARIABLE_NAME}}`
- 変数名は大文字と小文字が区別されます
- 変数は**必ず**アンダースコアでプレフィックスする必要があります（例: `{{_VARIABLE}}`）。これはテンプレートのインクルードと区別するためです
- ネストされたプロパティにはフォワードスラッシュでアクセスします: `{{_OBJECT/property}}`
- 配列要素はインデックスでアクセスします: `{{_ARRAY/0}}`

### 式
- 数学的表現: `{{1 + 2}}`, `{{1 + 2 * 3}}`, `{{(1 + 2) * 3}}`
- 比較式: `{{_X == "value"}}`, `{{_X != "value"}}`
- 論理演算子: `&&` (AND), `||` (OR), `!` (NOT)
- グループ化のための括弧: `{{(_X == "a") || (_Y == "b")}}`

## 関数

関数は `@` プレフィックスで呼び出します:

```
{{@function_name("param1", "param2")}}
```

### コア関数

#### @string(value)
任意の型の値を文字列表現に変換します。

**例:**
```
{{@string(123)}} は "123" を出力
{{@string(true)}} は "true" を出力
```

#### @printf(format, args...)
指定されたフォーマット文字列と引数を使用して文字列をフォーマットします。Go の fmt.Sprintf の規則に従います。

**例:**
```
{{@printf("こんにちは、%s！あなたは%d歳です。", "田中", 30)}}
は "こんにちは、田中！あなたは30歳です。" を出力
```

#### @error(format, args...)
フォーマットされたメッセージでエラーを生成します。try/catchブロックで捕捉されない限り、テンプレート処理が停止します。

**例:**
```
{{try}}
  {{@error("無効な値: %s", _VALUE)}}
{{catch _E}}
  エラーが発生しました: {{_E}}
{{/try}}
```

#### @redirect(url)
指定されたURLにリダイレクトします。これは通常、Webアプリケーションで使用されます。

**例:**
```
{{if _USER_LOGGED_OUT}}
  {{@redirect("/login")}}
{{/if}}
```

#### @rand(min, max)
min から max までの範囲（両端含む）のランダムな整数を生成します。

**例:**
```
1から10までのランダムな数: {{@rand(1, 10)}}
```

#### @seq(start, end, [step])
開始から終了までのオプションのステップ増分（デフォルトのステップは1）で数値のシーケンスを生成します。

**例:**
```
{{foreach {{@seq(1, 5)}} as _NUM}}
  {{_NUM}} 
{{/foreach}}
は "1 2 3 4 5" を出力

{{foreach {{@seq(0, 10, 2)}} as _NUM}}
  {{_NUM}} 
{{/foreach}}
は "0 2 4 6 8 10" を出力
```

## フィルター

フィルターは値を変換し、パイプ記号で適用されます:

```
{{variable|filter(params)}}
```

### 文字列フィルター

#### |uppercase()
文字列を大文字に変換します。

**例:**
```
{{"hello world"|uppercase()}} は "HELLO WORLD" を出力
```

#### |lowercase()
文字列を小文字に変換します。

**例:**
```
{{"HELLO WORLD"|lowercase()}} は "hello world" を出力
```

#### |trim()
文字列の先頭と末尾の空白を削除します。

**例:**
```
{{" hello world "|trim()}} は "hello world" を出力
```

#### |substr(start, length)
指定された位置から指定された長さの部分文字列を抽出します。

**例:**
```
{{"hello world"|substr(0, 5)}} は "hello" を出力
{{"hello world"|substr(6, 5)}} は "world" を出力
```

#### |truncate(length, [ellipsis], [word_cut])
文字列を指定された長さに切り詰め、指定された場合は省略記号を追加します。word_cutがfalseの場合、単語の境界で切り詰めます。

**例:**
```
{{"これは非常に長い文章です"|truncate(10)}} は "これは非常に..." を出力
{{"これは非常に長い文章です"|truncate(10, "...")}} は "これは非常に..." を出力
{{"これは非常に長い文章です"|truncate(10, "...", false)}} は "これは..." を出力
```

#### |entities()
HTML特殊文字をそのエンティティ等価物に変換します。

**例:**
```
{{"<div>Hello & World</div>"|entities()}} は "&lt;div&gt;Hello &amp; World&lt;/div&gt;" を出力
```

#### |striptags()
文字列からHTMLタグを削除します。

**例:**
```
{{"<div>Hello <strong>World</strong></div>"|striptags()}} は "Hello World" を出力
```

#### |replace(from, to)
部分文字列のすべての出現を別の部分文字列に置き換えます。

**例:**
```
{{"hello world"|replace("world", "universe")}} は "hello universe" を出力
```

#### |nl2br()
改行（\n）をHTML改行タグ（<br/>）に変換します。

**例:**
```
{{"Line 1\nLine 2"|nl2br()}} は "Line 1<br/>Line 2" を出力
```

### 配列フィルター

#### |json()
値をそのJSON表現に変換します。

**例:**
```
{{set _ARRAY=(1,2,3,4,5)}}
{{_ARRAY|json()}} は "[1,2,3,4,5]" を出力

{{set _OBJ={"name": "田中", "age": 30}}}
{{_OBJ|json()}} は "{"name":"田中","age":30}" を出力
```

#### |jsonparse()
JSON文字列をオブジェクトまたは配列に解析します。

**例:**
```
{{set _JSON_STR='{"name":"田中","age":30}'}}
{{set _OBJ=_JSON_STR|jsonparse()}}
{{_OBJ/name}} は "田中" を出力
```

#### |explode(delimiter)
指定された区切り文字を使用して文字列を配列に分割します。

**例:**
```
{{set _PARTS="りんご,オレンジ,バナナ"|explode(",")}}
{{_PARTS/0}} は "りんご" を出力
{{_PARTS/1}} は "オレンジ" を出力
```

#### |implode(glue)
配列要素を指定された接着文字列で結合します。

**例:**
```
{{set _ARRAY=("りんご", "オレンジ", "バナナ")}}
{{_ARRAY|implode(", ")}} は "りんご, オレンジ, バナナ" を出力
```

#### |reverse()
配列または文字列を反転します。

**例:**
```
{{set _ARRAY=("りんご", "オレンジ", "バナナ")}}
{{_ARRAY|reverse()|implode(", ")}} は "バナナ, オレンジ, りんご" を出力

{{"こんにちは"|reverse()}} は "はちにんこ" を出力
```

#### |arrayslice(offset, [length])
指定されたオフセットから始まるオプションの長さの配列の一部を抽出します。

**例:**
```
{{set _ARRAY=("りんご", "オレンジ", "バナナ", "ぶどう", "キウイ")}}
{{_ARRAY|arrayslice(1, 2)|implode(", ")}} は "オレンジ, バナナ" を出力
```

#### |arrayfilter(path, value)
指定されたパスが指定された値と等しい要素のみを保持することで配列をフィルタリングします。

**例:**
```
{{set _USERS=(
  {"name": "田中", "age": 30, "active": true},
  {"name": "佐藤", "age": 25, "active": false},
  {"name": "鈴木", "age": 40, "active": true}
)}}
{{_USERS|arrayfilter("active", true)|json()}}
は '[{"name":"田中","age":30,"active":true},{"name":"鈴木","age":40,"active":true}]' を出力
```

#### |length() または |count()
配列または文字列の長さを取得します。

**例:**
```
{{set _ARRAY=("りんご", "オレンジ", "バナナ")}}
{{_ARRAY|length()}} は "3" を出力

{{"こんにちは"|length()}} は "5" を出力
```

#### |lines(count)
配列アイテムを指定された行あたりのアイテム数で行にグループ化します。

**例:**
```
{{set _ITEMS=("item1", "item2", "item3", "item4", "item5", "item6")}}
{{foreach {{_ITEMS|lines(2)}} as _LINE}}
  <div class="row">
  {{foreach {{_LINE}} as _ITEM}}
    <div class="col">{{_ITEM}}</div>
  {{/foreach}}
  </div>
{{/foreach}}
```

#### |columns(count)
配列アイテムを指定された列あたりのアイテム数で列にグループ化します。

**例:**
```
{{set _ITEMS=("item1", "item2", "item3", "item4", "item5", "item6")}}
{{foreach {{_ITEMS|columns(2)}} as _COL}}
  <div class="column">
  {{foreach {{_COL}} as _ITEM}}
    <div class="item">{{_ITEM}}</div>
  {{/foreach}}
  </div>
{{/foreach}}
```

### 型変換フィルター

#### |toint()
値を整数に変換します。

**例:**
```
{{"123"|toint()}} は 123 を出力
{{"123.45"|toint()}} は 123 を出力
```

#### |tostring()
値を文字列に変換します。@string関数と同様です。

**例:**
```
{{123|tostring()}} は "123" を出力
{{true|tostring()}} は "true" を出力
```

#### |round([precision])
数値を指定された小数精度に丸めます（デフォルトは0）。

**例:**
```
{{3.14159|round()}} は 3 を出力
{{3.14159|round(2)}} は 3.14 を出力
{{3.14159|round(4)}} は 3.1416 を出力
```

### エンコーディングフィルター

#### |b64enc()
文字列をBase64にエンコードします。

**例:**
```
{{"こんにちは"|b64enc()}} は "44GT44KT44Gr44Gh44Gv" を出力
```

#### |b64dec()
Base64エンコードされた文字列をデコードします。

**例:**
```
{{"44GT44KT44Gr44Gh44Gv"|b64dec()}} は "こんにちは" を出力
```

#### |urlencode()
文字列をURLエンコードし、スペースをプラス記号に置き換えます。

**例:**
```
{{"こんにちは 世界"|urlencode()}} は "%E3%81%93%E3%82%93%E3%81%AB%E3%81%A1%E3%81%AF+%E4%B8%96%E7%95%8C" を出力
{{"a=b&c=d"|urlencode()}} は "a%3Db%26c%3Dd" を出力
```

#### |rawurlencode()
文字列を生のURLエンコードし、スペースを%20として維持します。

**例:**
```
{{"こんにちは 世界"|rawurlencode()}} は "%E3%81%93%E3%82%93%E3%81%AB%E3%81%A1%E3%81%AF%20%E4%B8%96%E7%95%8C" を出力
```

### フォーマットフィルター

#### |size()
バイトサイズを人間が読みやすい形式（KB、MB、GBなど）にフォーマットします。

**例:**
```
{{1024|size()}} は "1 KB" を出力
{{1048576|size()}} は "1 MB" を出力
{{1073741824|size()}} は "1 GB" を出力
```

#### |duration()
秒単位の時間を人間が読みやすい時間形式にフォーマットします。

**例:**
```
{{3600|duration()}} は "1:00:00"（1時間）を出力
{{90|duration()}} は "1:30"（1分30秒）を出力
```

### 特殊フィルター

#### |dump() または |export()
変数に関するデバッグ情報を出力し、その型と値を表示します。

**例:**
```
{{set _OBJ={"name": "田中", "age": 30}}}
{{_OBJ|dump()}}
は以下のような出力をします:
[object map[string]interface {}: {"age":30,"name":"田中"}]
```

#### |type()
変数の型を文字列として返します。

**例:**
```
{{123|type()}} は "int" を出力
{{"こんにちは"|type()}} は "string" を出力
{{(1,2,3)|type()}} は "array" を出力
```

## 制御構造

### 条件文
```
{{if condition}}
  条件が真の場合のコンテンツ
{{elseif other_condition}}
  他の条件が真の場合のコンテンツ
{{else}}
  すべての条件が偽の場合のコンテンツ
{{/if}}
```

### ループ
```
{{foreach {{array}} as _item}}
  アクセス: {{_item}}
  インデックス: {{_item_idx}}
  キー: {{_item_key}}
  最大: {{_item_max}}
  前の項目: {{_item_prv}}
{{else}}
  配列が空の場合に表示されるコンテンツ
{{/foreach}}
```

### 変数代入
```
{{set _X="value"}}
  {{_X}}が設定されたコンテンツ
{{/set}}

{{set _X="value" _Y="another value"}}
  複数の変数: {{_X}}と{{_Y}}
{{/set}}

{{set _X=(1 + 2)}}
  式の結果: {{_X}}
{{/set}}
```

### エラー処理
```
{{try}}
  {{@error("何か問題が発生しました")}}
{{catch _E}}
  エラーが発生しました: {{_E}}
{{/try}}
```

## 変数とスコープ

- `{{set}}`で定義された変数はブロック内で利用可能です
- 特殊変数:
  - `{{_TPL_PAGE}}` - 現在のテンプレートページ
  - `{{_EXCEPTION}}` - try/catchブロック内のエラー

## 特殊機能

### リテラルテキスト
```
{{literal}}
  処理されない{{tags}}を含むコンテンツ
{{/literal}}
```

### フォーマット変換
- Markdown から HTML: `{{content|markdown()}}`
- BBCode から HTML: `{{content|bbcode()}}`

### 関数結果のチェーン
```
{{@string("テキスト")|uppercase()|json()}}
```

### シンプルなForeachの簡潔な構文
```
{{foreach (1,2,3) as _X}}{{_X}}{{/foreach}}
```

### 複雑なオブジェクトのループ処理
ループ内でネストされたプロパティにアクセスします:
```
{{foreach {{_COMPLEX_OBJECT}} as _X}}
  {{if {{_X/property}}=="value"}}
    {{_X/property}}
  {{/if}}
{{/foreach}}
```