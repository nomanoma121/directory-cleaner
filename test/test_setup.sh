#!/bin/bash

set -e  # エラーで即終了

# 各テストディレクトリと使用するパッケージマネージャの設定
declare -A dirs=(
  ["test1"]="."
  ["test2"]="tmp/tmp"
  ["test3"]="."
  ["test4"]="."       # yarn
  ["test5"]="."       # pnpm
)

declare -A managers=(
  ["test1"]="npm"
  ["test2"]="npm"
  ["test3"]="npm"
  ["test4"]="yarn"
  ["test5"]="pnpm"
)

for test in "${!dirs[@]}"; do
  subdir="${dirs[$test]}"
  target="$test/$subdir"
  manager="${managers[$test]}"

  echo "==> セットアップ中: $target （使用: $manager）"

  if [ -d "$target/node_modules" ]; then
    echo "❌ $target はすでにセットアップされています。スキップします。"
    continue
  fi

  # ディレクトリ構成作成
  mkdir -p "$target"
  touch "$test/tmp.md"

  # package.json が無ければ作成
  if [ ! -f "$target/package.json" ]; then
    echo "{}" > "$target/package.json"
  fi

  # 適切なコマンドで Next.js をインストール
  case "$manager" in
    npm)
      (cd "$target" && npm install next)
      ;;
    yarn)
      (cd "$target" && yarn add next)
      ;;
    pnpm)
      (cd "$target" && pnpm add next)
      ;;
    *)
      echo "⚠️ 未知のパッケージマネージャ: $manager"
      exit 1
      ;;
  esac

  echo "✅ $target のセットアップ完了！"
done
