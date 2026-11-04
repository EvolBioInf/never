databases="/home/cloud/dbs/neidb/web/databases"

curr=$(date -d 2026-02-01 +%b)

months=("Jan" "Feb" "Mar" "Apr" "May" "Jun" "Jul" "Aug" "Sep" "Oct" "Nov" "Dec")
declare -A rev_months=(
  [Jan]=0
  [Feb]=1
  [Mar]=2
  [Apr]=3
  [May]=4
  [Jun]=5
  [Jul]=6
  [Aug]=7
  [Sep]=8
  [Oct]=9
  [Nov]=10
  [Dec]=11
)

month_idx=-1
for i in {0..11}; do
  if [ "$curr" = ${months[$i]} ]; then
    month_idx=$i
  fi
done

# We calculate a score out of the year and month a database was constructed
# The score is year * 100 + month
# With this scoring term, we calculate a threshold, which is six months past.
# Every filename scoring below/ equal the threshold gets deleted.

year=$(date +%Y)
month_thresh=$((($month_idx) - 6))
if (( $month_thresh < 0 )); then
  year=$((($year) - 1))
  month_thresh=$((( ${rev_months[Dec]} ) + ( $month_thresh )))
fi

threshold=$(($year*100 + $month_thresh))

echo "Deletion threshold" $year\_${months[$month_thresh]}
echo "Removing the following files: "
for filepath in "$databases"/*; do
  b=$(basename $filepath)
  if [[ "$b" =~ ^temp_([0-9]{4})_([A-Z][a-z]{2}).* ]]; then
    score=$(( ${BASH_REMATCH[1]} * 100 + ${rev_months[${BASH_REMATCH[2]}]}))
    if (( $score <= $threshold )); then
      echo $b
      rm $filepath
    fi
  fi
done
