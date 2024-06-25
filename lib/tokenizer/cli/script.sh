curl -s https://ftp.uniprot.org/pub/databases/uniprot/current_release/knowledgebase/complete/uniprot_sprot.xml.gz | gzip -d -k -c | go run main.go --outputDir output
