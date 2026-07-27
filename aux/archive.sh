DB_NAME=$(date +%Y_%b_%d)
echo "archiving neidb as ${DB_NAME}.db and ${DB_NAME}.tgz"
cp ../data/neidb ../never/databases/$DB_NAME.db
cp ../data/neidbData.tgz ../never/databases/$DB_NAME.tgz
