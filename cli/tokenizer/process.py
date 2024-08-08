import os
import sqlite3
import numpy as np
from tqdm import tqdm

# Connection to your database
db_path = "./sequences.db"
conn = sqlite3.connect(db_path)

# Calculate split index for training and validation
def calculate_split_index(total_rows, val_percentage):
    return int(total_rows * (1 - val_percentage))

def fetch_data(val_percentage=0.01):
    cursor = conn.cursor()
    cursor.execute("SELECT COUNT(*) FROM sequences")
    total_rows = cursor.fetchone()[0]
    split_index = calculate_split_index(total_rows, val_percentage)

    # Fetch data with randomized order
    cursor.execute("SELECT tokens FROM sequences ORDER BY RANDOM()")

    count = 0
    while True:
        row = cursor.fetchone()
        if row is None:
            break
        yield row[0], count < split_index
        count += 1

    cursor.close()

# Function to convert blob bytes to uint16 array
def bytes_to_uint16(buf):
    arr = np.frombuffer(buf, dtype=np.uint16)
    return np.append(arr, 0)  # Append 0 as the EOT token

if __name__ == '__main__':
    train_filename = os.path.join(os.path.dirname(__file__), 'train.bin')
    val_filename = os.path.join(os.path.dirname(__file__), 'val.bin')
    dtype = np.uint16

    # Initialize memmap files with rough size estimates, adjusted as needed
    train_arr = np.memmap(train_filename, dtype=dtype, mode='w+', shape=(1,))
    val_arr = np.memmap(val_filename, dtype=dtype, mode='w+', shape=(1,))

    train_idx = 0
    val_idx = 0
    for tokens, is_train in fetch_data():
        tokens_uint16 = bytes_to_uint16(tokens)

        # Determine where to store the tokens
        if is_train:
            if train_idx + len(tokens_uint16) > len(train_arr):
                train_arr.flush()
                train_arr = np.memmap(train_filename, dtype=dtype, mode='r+', shape=(train_idx + len(tokens_uint16),))
            train_arr[train_idx:train_idx + len(tokens_uint16)] = tokens_uint16
            train_idx += len(tokens_uint16)
        else:
            if val_idx + len(tokens_uint16) > len(val_arr):
                val_arr.flush()
                val_arr = np.memmap(val_filename, dtype=dtype, mode='r+', shape=(val_idx + len(tokens_uint16),))
            val_arr[val_idx:val_idx + len(tokens_uint16)] = tokens_uint16
            val_idx += len(tokens_uint16)

    train_arr.flush()
    val_arr.flush()
    conn.close()

    print(f"Training data written to {train_filename}. Size: {train_idx * np.dtype(dtype).itemsize / (1024**2)} MB")
    print(f"Validation data written to {val_filename}. Size: {val_idx * np.dtype(dtype).itemsize / (1024**2)} MB")

