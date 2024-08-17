import timeit
import os
from Bio import SeqIO
from dnadesign.parsers import parse_genbank_from_c_file

def benchmark_dnadesign(file_path, num_iterations=10):
    def parse_with_dnadesign():
        records = parse_genbank_from_c_file(file_path)
        return records

    time_taken = timeit.timeit(parse_with_dnadesign, number=num_iterations)
    return time_taken / num_iterations

def benchmark_biopython(file_path, num_iterations=10):
    def parse_with_biopython():
        with open(file_path, "r") as handle:
            records = list(SeqIO.parse(handle, "genbank"))
        return records

    time_taken = timeit.timeit(parse_with_biopython, number=num_iterations)
    return time_taken / num_iterations

def main():
    current_dir = os.path.dirname(__file__)
    example_path = os.path.join(current_dir, '../lib/bio/genbank/data/bsub.gbk')

    print(f"Benchmarking GenBank parsing for file: {example_path}")
    print("Running 10 iterations for each parser...")

    dnadesign_time = benchmark_dnadesign(example_path)
    biopython_time = benchmark_biopython(example_path)

    print(f"\nResults:")
    print(f"DNA design average time: {dnadesign_time:.6f} seconds")
    print(f"BioPython average time:  {biopython_time:.6f} seconds")

    speedup = biopython_time / dnadesign_time
    print(f"\nDNA design is {speedup:.2f}x faster than BioPython")

if __name__ == "__main__":
    main()
