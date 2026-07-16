PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE taxon (
            taxid int primary key,
            parent int,
            name text,
            common_name text,
            rank text,
            score float);
INSERT INTO taxon VALUES(89593,7711,'Craniata','','subphylum',14153.0);
INSERT INTO taxon VALUES(131567,1,'cellular organisms','','cellular root',3332975.0);
INSERT INTO taxon VALUES(2759,131567,'Eukaryota','eukaryotes','domain',3702751.0);
INSERT INTO taxon VALUES(32523,1338369,'Tetrapoda','tetrapods','clade',3646407.0);
INSERT INTO taxon VALUES(8287,117571,'Sarcopterygii','','superclass',10092.0);
INSERT INTO taxon VALUES(1,1,'root','','no rank',3636325.0);
INSERT INTO taxon VALUES(41666,8292,'Batrachia','','superorder',334.0);
INSERT INTO taxon VALUES(33511,33213,'Deuterostomia','','clade',14475.0);
INSERT INTO taxon VALUES(8292,32523,'Amphibia','amphibians','class',3636667.0);
INSERT INTO taxon VALUES(6072,33208,'Eumetazoa','','clade',28694.0);
INSERT INTO taxon VALUES(117570,7776,'Teleostomi','','clade',13986.0);
INSERT INTO taxon VALUES(8342,41666,'Anura','frogs & toads','order',3636587.0);
INSERT INTO taxon VALUES(33213,6072,'Bilateria','','clade',28158.0);
INSERT INTO taxon VALUES(7776,7742,'Gnathostomata','jawed vertebrates','clade',3650442.0);
INSERT INTO taxon VALUES(33208,33154,'Metazoa','animals','kingdom',3665244.0);
INSERT INTO taxon VALUES(33154,2759,'Opisthokonta','','clade',54230.0);
INSERT INTO taxon VALUES(8397,30352,'Ranidae','riparian frogs','family',3636356.0);
INSERT INTO taxon VALUES(8399,8397,'Rana','','genus',9.0);
INSERT INTO taxon VALUES(1338369,8287,'Dipnotetrapodomorpha','','clade',10088.0);
INSERT INTO taxon VALUES(8407,121175,'Rana temporaria','common frog','species',3636329.0);
INSERT INTO taxon VALUES(121175,8399,'Rana','','subgenus',9.0);
INSERT INTO taxon VALUES(7711,33511,'Chordata','chordates','phylum',3650607.0);
INSERT INTO taxon VALUES(7742,89593,'Vertebrata','vertebrates','clade',3650478.0);
INSERT INTO taxon VALUES(117571,117570,'Euteleostomi','bony vertebrates','clade',3650311.0);
INSERT INTO taxon VALUES(30352,8416,'Ranoidea','','superfamily',59.0);
INSERT INTO taxon VALUES(8416,8342,'Neobatrachia','','suborder',202.0);
CREATE TABLE genome (
            taxid int,
            size real, 
            accession text primary key,
            level text,
            foreign key(taxid) references taxon(taxid));
INSERT INTO genome VALUES(8407,4282.72462600000017,'GCA_009802015.1','scaffold');
INSERT INTO genome VALUES(8407,4039.80289600000014,'GCA_058280865.1','chromosome');
INSERT INTO genome VALUES(8407,3686.21012900000005,'GCA_905171725.1','scaffold');
INSERT INTO genome VALUES(8407,4111.42259600000034,'GCF_905171775.1','chromosome');
CREATE TABLE genome_count (
            taxid int,
            level text,
            raw int,
            recursive int,
            primary key(taxid, level),
            foreign key(taxid) references taxon(taxid));
INSERT INTO genome_count VALUES(117571,'complete',0,80);
INSERT INTO genome_count VALUES(117571,'chromosome',0,4240);
INSERT INTO genome_count VALUES(117571,'scaffold',0,7984);
INSERT INTO genome_count VALUES(117571,'contig',0,1682);
INSERT INTO genome_count VALUES(30352,'complete',0,0);
INSERT INTO genome_count VALUES(30352,'chromosome',0,20);
INSERT INTO genome_count VALUES(30352,'scaffold',0,38);
INSERT INTO genome_count VALUES(30352,'contig',0,1);
INSERT INTO genome_count VALUES(8416,'complete',0,0);
INSERT INTO genome_count VALUES(8416,'chromosome',0,56);
INSERT INTO genome_count VALUES(8416,'scaffold',0,139);
INSERT INTO genome_count VALUES(8416,'contig',0,7);
INSERT INTO genome_count VALUES(89593,'complete',0,80);
INSERT INTO genome_count VALUES(89593,'chromosome',0,4317);
INSERT INTO genome_count VALUES(89593,'scaffold',0,8063);
INSERT INTO genome_count VALUES(89593,'contig',0,1693);
INSERT INTO genome_count VALUES(131567,'complete',0,86797);
INSERT INTO genome_count VALUES(131567,'chromosome',0,26889);
INSERT INTO genome_count VALUES(131567,'scaffold',0,564262);
INSERT INTO genome_count VALUES(131567,'contig',0,2655027);
INSERT INTO genome_count VALUES(2759,'complete',0,1173);
INSERT INTO genome_count VALUES(2759,'chromosome',0,16150);
INSERT INTO genome_count VALUES(2759,'scaffold',0,32340);
INSERT INTO genome_count VALUES(2759,'contig',0,16763);
INSERT INTO genome_count VALUES(32523,'complete',0,35);
INSERT INTO genome_count VALUES(32523,'chromosome',0,2971);
INSERT INTO genome_count VALUES(32523,'scaffold',0,5710);
INSERT INTO genome_count VALUES(32523,'contig',0,1366);
INSERT INTO genome_count VALUES(8287,'complete',0,35);
INSERT INTO genome_count VALUES(8287,'chromosome',0,2976);
INSERT INTO genome_count VALUES(8287,'scaffold',0,5713);
INSERT INTO genome_count VALUES(8287,'contig',0,1368);
INSERT INTO genome_count VALUES(1,'complete',0,321836);
INSERT INTO genome_count VALUES(1,'chromosome',0,47724);
INSERT INTO genome_count VALUES(1,'scaffold',0,573235);
INSERT INTO genome_count VALUES(1,'contig',0,2693530);
INSERT INTO genome_count VALUES(41666,'complete',0,0);
INSERT INTO genome_count VALUES(41666,'chromosome',0,96);
INSERT INTO genome_count VALUES(41666,'scaffold',0,225);
INSERT INTO genome_count VALUES(41666,'contig',0,13);
INSERT INTO genome_count VALUES(33511,'complete',0,83);
INSERT INTO genome_count VALUES(33511,'chromosome',0,4463);
INSERT INTO genome_count VALUES(33511,'scaffold',0,8208);
INSERT INTO genome_count VALUES(33511,'contig',0,1721);
INSERT INTO genome_count VALUES(8292,'complete',0,0);
INSERT INTO genome_count VALUES(8292,'chromosome',0,100);
INSERT INTO genome_count VALUES(8292,'scaffold',0,229);
INSERT INTO genome_count VALUES(8292,'contig',0,13);
INSERT INTO genome_count VALUES(6072,'complete',0,140);
INSERT INTO genome_count VALUES(6072,'chromosome',0,9597);
INSERT INTO genome_count VALUES(6072,'scaffold',0,13676);
INSERT INTO genome_count VALUES(6072,'contig',0,5281);
INSERT INTO genome_count VALUES(117570,'complete',0,80);
INSERT INTO genome_count VALUES(117570,'chromosome',0,4240);
INSERT INTO genome_count VALUES(117570,'scaffold',0,7984);
INSERT INTO genome_count VALUES(117570,'contig',0,1682);
INSERT INTO genome_count VALUES(8342,'complete',0,0);
INSERT INTO genome_count VALUES(8342,'chromosome',0,87);
INSERT INTO genome_count VALUES(8342,'scaffold',0,165);
INSERT INTO genome_count VALUES(8342,'contig',0,10);
INSERT INTO genome_count VALUES(33213,'complete',0,138);
INSERT INTO genome_count VALUES(33213,'chromosome',0,9457);
INSERT INTO genome_count VALUES(33213,'scaffold',0,13394);
INSERT INTO genome_count VALUES(33213,'contig',0,5169);
INSERT INTO genome_count VALUES(7776,'complete',0,80);
INSERT INTO genome_count VALUES(7776,'chromosome',0,4302);
INSERT INTO genome_count VALUES(7776,'scaffold',0,8046);
INSERT INTO genome_count VALUES(7776,'contig',0,1689);
INSERT INTO genome_count VALUES(33208,'complete',0,140);
INSERT INTO genome_count VALUES(33208,'chromosome',0,9685);
INSERT INTO genome_count VALUES(33208,'scaffold',0,13775);
INSERT INTO genome_count VALUES(33208,'contig',0,5319);
INSERT INTO genome_count VALUES(33154,'complete',0,839);
INSERT INTO genome_count VALUES(33154,'chromosome',0,11287);
INSERT INTO genome_count VALUES(33154,'scaffold',0,27608);
INSERT INTO genome_count VALUES(33154,'contig',0,14496);
INSERT INTO genome_count VALUES(8397,'complete',0,0);
INSERT INTO genome_count VALUES(8397,'chromosome',0,13);
INSERT INTO genome_count VALUES(8397,'scaffold',0,18);
INSERT INTO genome_count VALUES(8397,'contig',0,0);
INSERT INTO genome_count VALUES(8399,'complete',0,0);
INSERT INTO genome_count VALUES(8399,'chromosome',0,7);
INSERT INTO genome_count VALUES(8399,'scaffold',0,2);
INSERT INTO genome_count VALUES(8399,'contig',0,0);
INSERT INTO genome_count VALUES(1338369,'complete',0,35);
INSERT INTO genome_count VALUES(1338369,'chromosome',0,2975);
INSERT INTO genome_count VALUES(1338369,'scaffold',0,5710);
INSERT INTO genome_count VALUES(1338369,'contig',0,1368);
INSERT INTO genome_count VALUES(8407,'complete',0,0);
INSERT INTO genome_count VALUES(8407,'chromosome',2,2);
INSERT INTO genome_count VALUES(8407,'scaffold',2,2);
INSERT INTO genome_count VALUES(8407,'contig',0,0);
INSERT INTO genome_count VALUES(121175,'complete',0,0);
INSERT INTO genome_count VALUES(121175,'chromosome',0,7);
INSERT INTO genome_count VALUES(121175,'scaffold',0,2);
INSERT INTO genome_count VALUES(121175,'contig',0,0);
INSERT INTO genome_count VALUES(7711,'complete',0,80);
INSERT INTO genome_count VALUES(7711,'chromosome',0,4382);
INSERT INTO genome_count VALUES(7711,'scaffold',0,8110);
INSERT INTO genome_count VALUES(7711,'contig',0,1710);
INSERT INTO genome_count VALUES(7742,'complete',0,80);
INSERT INTO genome_count VALUES(7742,'chromosome',0,4317);
INSERT INTO genome_count VALUES(7742,'scaffold',0,8063);
INSERT INTO genome_count VALUES(7742,'contig',0,1693);
CREATE TABLE image (
             image_id int,
             url text,
             attribution text,
             primary key(image_id));
INSERT INTO image VALUES(67186,'http://www.ncbi.nlm.nih.gov/Taxonomy/taxi/images/2810','Richard Bartz');
CREATE TABLE tax2ima (
             taxid int,
             image_id int,
             primary key(taxid, image_id),
             foreign key(taxid) references taxon(taxid),
             foreign key(image_id) references image(image_id));
INSERT INTO tax2ima VALUES(8407,67186);
CREATE INDEX taxon_parent_idx on taxon(parent);
CREATE INDEX taxon_score_idx on taxon(score);
CREATE INDEX genome_taxid_idx on genome(taxid);
CREATE INDEX genome_size_idx on genome(size);
CREATE INDEX genome_count_raw_idx on genome_count(raw);
CREATE INDEX genome_count_recursive_idx on
            genome_count(recursive);
CREATE INDEX tax2ima_taxid_idx on tax2ima(taxid);
CREATE INDEX tax2ima_image_id_idx on tax2ima(image_id);
COMMIT;
