#!/bin/bash

set -e  # エラーで即終了

# 各テストディレクトリの設定
declare -A dirs=(
  ["test1"]="."
  ["test2"]="tmp/tmp"
  ["test3"]="."
)

for test in "${!dirs[@]}"; do
  subdir="${dirs[$test]}"
  target="$test/$subdir"

  echo "==> セットアップ中: $target"

  if [ -d "$target/node_modules" ]; then
    echo "❌ $target はすでにセットアップされています。スキップします。"
    continue
  fi

  # ディレクトリ構成作成
  mkdir -p "$target"
  touch "$test/tmp.md"

  # npm init (必要であれば package.json を作る)
  if [ ! -f "$target/package.json" ]; then
    echo "{}" > "$target/package.json"
  fi

  # Next.js をインストール
  (cd "$target" && npm install next)

  echo "✅ $target のセットアップ完了！"
done
