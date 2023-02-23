for c in contracts/*; do
    (cd $c && cargo schema)
done
