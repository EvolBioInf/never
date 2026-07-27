neverRoot="/home/cloud/dbs/neidb/web"
DB_NAME=$(date +%Y_%b_%d)
echo "archiving neidb as ${DB_NAME}.db and ${DB_NAME}.tgz"
cp $neverRoot/../neidb $neverRoot/databases/$DB_NAME.db
cp $neverRoot/data/neidbData.tgz $neverRoot/databases/$DB_NAME.tgz
