# 環境変数

このアプリケーションでは、環境変数によっていくつかの挙動変更が可能です。

設定ファイルで設定できる項目は、環境変数でも設定が可能です。該当する設定が環境変数でも設定されている場合、環境変数の値が優先されます。  
このページでは、設定ファイルで設定できる項目について説明しておりません。[設定ファイル](file.md) をご覧ください。

## 設定項目

環境変数では、設定ファイルで設定できる項目に加え、以下の設定変更が可能です。

- `CONFIG_PATH`: 設定ファイルのパス

### CONFIG_PATH

| 必須か | デフォルト値 |
| :- | :- |
| いいえ | `data/config.yml` |

設定ファイルのパスを設定します。  
このアプリケーションは、既定では実行ファイルと同じ階層の `data` フォルダにある `config.yml` を設定ファイルとして使用します。
