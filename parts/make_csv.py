import os
import yaml
import csv
import sys

def process_yaml_files(directory):
    data = []
    for filename in os.listdir(directory):
        if filename.endswith('.yaml'):
            with open(os.path.join(directory, filename), 'r') as file:
                yaml_data = yaml.safe_load(file)
                for name, gene in yaml_data.items():
                    sequence = gene['prefix'] + gene['sequence'].lower() + gene['suffix']
                    data.append({
                        'name': name,
                        'vector': 'pOpen_v3',
                        'type': 'dna',
                        'sequence': sequence
                    })
    return data

def print_csv(data):
    writer = csv.DictWriter(sys.stdout, fieldnames=['name', 'vector', 'type', 'sequence'])
    writer.writeheader()
    for row in data:
        writer.writerow(row)

if __name__ == '__main__':
    directory = './parts'
    yaml_data = process_yaml_files(directory)
    print_csv(yaml_data)
