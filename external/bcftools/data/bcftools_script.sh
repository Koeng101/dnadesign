minimap2 -ax map-ont template.fasta reads.fastq.gz | samtools view -bS - | samtools sort | bcftools mpileup -Ou -f template.fasta - | bcftools call -mv -Ov -
