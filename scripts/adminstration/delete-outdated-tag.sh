#!/bin/bash

git fetch

existingfeaturebranches=$(git branch -r | sed 's/^ *//;s/ *$//' | egrep -v "(^\*|main)")

for tag in $(git tag --list 'v[0-9]*\.[0-9]*\.[0-9]*-*')
do
  # regex to match the tag name
  [[ $tag =~ ^v[0-9]*\.[0-9]*\.[0-9]*-(.*)$ ]]
  featurebranchname=${BASH_REMATCH[1]}
  echo "Feature branch name: $featurebranchname"
  if [[ $existingfeaturebranches =~ "feature/$featurebranchname" ]]
  then
    echo "Branch $featurebranchname exists, so not deleting tag $tag"
  else
    echo "Branch $featurebranchname does not exist, so deleting tag $tag"
    git push origin --delete $tag
  fi
done