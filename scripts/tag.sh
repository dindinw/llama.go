
echo "📌  go to llama.cpp submodule"
cd core/llama.cpp
pwd

tag=$1

if [ -z $tag ]; then
  # get the latest tag 
  last_tag=$(git ls-remote --tags origin "refs/tags/b*"|tail -n 1|cut -d/ -f3)
  tag=$last_tag
  echo "  ✔ the last remote tag is $tag"
else
  echo "  ✔  use a input manually tag: $tag"
fi


# if local exist 
if git show-ref --tags --quiet "refs/tags/$tag"; then
  echo "  ✔ Local already has tag: $tag (skip fetch)"
else
  echo "  → Fetching $tag from origin..."
  git fetch origin $tag 
fi

# checkout $tag
echo "📌  Checkout $tag"
git checkout $tag

cd - 
echo "✅  Done: $tag ready             "
