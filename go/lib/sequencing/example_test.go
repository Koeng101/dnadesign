package sequencing_test

//func Example_ampliconAlignment() {
//	// This is currently a work-in-progress. Sequencing utilities are under
//	// development right now.
//	//
//	//
//	// Only run function if minimap2 is available
//	_, err := exec.LookPath("minimap2")
//	if err != nil {
//		fmt.Println("oligo2")
//		return
//	}
//	// First, let's define the type we are looking for: amplicons in a pool.
//	type Amplicon struct {
//		Identifier       string
//		TemplateSequence string
//		ForwardPrimer    string
//		ReversePrimer    string
//	}
//
//	// Next, let's define data we'll be working on. In particular, the
//	// templates and fastq files.
//
//	/*
//		Data processing steps:
//
//		1. Simulate PCRs of amplicons
//		2. Sort for the right barcodes
//		3. Trim fastq reads
//		4. Minimap2 fastq reads to amplicons
//		5. Filter for primary alignments
//	*/
//	var amplicons []Amplicon
//	var templates []fasta.Record
//	pcrTm := 50.0
//
//	forward := "CCGTGCGACAAGATTTCAAG"
//	reverse := transform.ReverseComplement("CGGATCGAACTTAGGTAGCC")
//	oligo1 := Amplicon{Identifier: "oligo1", ForwardPrimer: forward, ReversePrimer: reverse, TemplateSequence: "CCGTGCGACAAGATTTCAAGGGTCTCTGTCTCAATGACCAAACCAACGCAAGTCTTAGTTCGTTCAGTCTCTATTTTATTCTTCATCACACTGTTGCACTTGGTTGTTGCAATGAGATTTCCTAGTATTTTCACTGCTGTGCTGAGACCCGGATCGAACTTAGGTAGCCT"}
//	oligo2 := Amplicon{Identifier: "oligo2", ForwardPrimer: forward, ReversePrimer: reverse, TemplateSequence: "CCGTGCGACAAGATTTCAAGGGTCTCTGTGCTATTTGCCGCTAGTTCCGCTCTAGCTGCTCCAGTTAATACTACTACTGAAGATGAATTGGAGGGTGACTTCGATGTTGCTGTTCTGCCTTTTTCCGCTTCTGAGACCCGGATCGAACTTAGGTAGCCACTAGTCATAAT"}
//	oligo3 := Amplicon{Identifier: "oligo3", ForwardPrimer: forward, ReversePrimer: reverse, TemplateSequence: "CCGTGCGACAAGATTTCAAGGGTCTCTCTTCTATCGCAGCCAAGGAAGAAGGTGTATCTCTAGAGAAGCGTCGAGTGAGACCCGGATCGAACTTAGGTAGCCCCCTTCGAAGTGGCTCTGTCTGATCCTCCGCGGATGGCGACACCATCGGACTGAGGATATTGGCCACA"}
//	amplicons = []Amplicon{oligo1, oligo2, oligo3}
//
//	// Simulate PCRs
//	for _, amplicon := range amplicons {
//		fragments, _ := pcr.Simulate([]string{amplicon.TemplateSequence}, pcrTm, false, []string{amplicon.ForwardPrimer, amplicon.ReversePrimer})
//		if len(fragments) != 1 {
//			log.Fatalf("Should only get 1 fragment from PCR!")
//		}
//		// In case your template will have multiple fragments
//		for _, fragment := range fragments {
//			// Make sure to reset identifier if you have more than 1 fragment.
//			templates = append(templates, fasta.Record{Identifier: amplicon.Identifier, Sequence: fragment})
//		}
//	}
//	var buf bytes.Buffer
//	for _, template := range templates {
//		_, _ = template.WriteTo(&buf)
//	}
//
//	// Trim fastq reads. All the following processes (trimming, minimap2,
//	// filtering) are all done concurrently.
//
//	// Setup barcodes and fastq files
//	barcode := "barcode06"
//	r, _ := os.Open("data/reads.fastq")
//	parser := bio.NewFastqParser(r)
//
//	// Setup errorGroups and channels
//	ctx := context.Background()
//	errorGroup, ctx := errgroup.WithContext(ctx)
//
//	fastqReads := make(chan fastq.Read)
//	fastqBarcoded := make(chan fastq.Read)
//	samReads := make(chan sam.Alignment)
//	samPrimary := make(chan sam.Alignment)
//
//	// Read fastqs into channel
//	errorGroup.Go(func() error {
//		return parser.ParseToChannel(ctx, fastqReads, false)
//	})
//
//	// Filter the right barcode fastqs from channel
//	errorGroup.Go(func() error {
//		return bio.FilterData(ctx, fastqReads, fastqBarcoded, func(data fastq.Read) bool { return data.Optionals["barcode"] == barcode })
//	})
//
//	// Run minimap
//	errorGroup.Go(func() error {
//		return minimap2.Minimap2Channeled(&buf, fastqBarcoded, samReads)
//	})
//
//	// Sort out primary alignments
//	errorGroup.Go(func() error {
//		return bio.FilterData(ctx, samReads, samPrimary, sam.Primary)
//	})
//
//	// Read all them alignments out into memory
//	var outputAlignments []sam.Alignment
//	for alignment := range samPrimary {
//		outputAlignments = append(outputAlignments, alignment)
//	}
//
//	fmt.Println(outputAlignments[0].RNAME)
//	// Output: oligo2
//}
