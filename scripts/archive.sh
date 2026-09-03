neverRoot="/home/cloud/dbs/neidb/web"

DB_NAME=$(date +%Y_%b_%d)

if [ "$(date +%d_%b)" != "01_Jan" ]; then
  DB_NAME=temp_$DB_NAME
fi

echo "archiving neidb as ${DB_NAME}.db and ${DB_NAME}.tgz"
cp $neverRoot/../neidb $neverRoot/databases/$DB_NAME.db
cp $neverRoot/data/neidbData.tgz $neverRoot/databases/$DB_NAME.tgz
