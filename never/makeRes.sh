prog="../bin/fetch"
url="http://localhost:8080"
$prog $url > r1.txt
q="?t=9606"
$prog "${url}/children$q" > r2.txt
q="?t=562"
$prog "${url}/num_genomes$q" > r3.txt
$prog "${url}/num_genomes_rec$q" > r4.txt
$prog "${url}/parent$q" > r5.txt
q="?t=9606"
$prog "${url}/subtree$q" > r6.txt
q="?t=278148,602633"
$prog "${url}/accessions$q" > r7.txt
q="?t=9606,741158,63221"
$prog "${url}/mrca$q" > r8.txt
q="?t=9606,9605"
$prog "${url}/names$q" > r9.txt
q="?t=9606,40674"
$prog "${url}/path$q" > r10.txt
q="?t=9606,9605"
$prog "${url}/ranks$q" > r11.txt
q="?t=562,9606"
$prog "${url}/taxa_info$q" > r12.txt
q="?t=homo+sapiens"
$prog "${url}/taxids$q" > r13.txt
q="?t=dolph&n=10&p=2"
$prog "${url}/taxi$q" > r14.txt
q="?a=GCF_000001405.40,GCA_000002115.2"
$prog "${url}/levels$q" > r15.txt
