-- 1. Tambahkan kolom user_id ke tabel account_types sebagai nullable terlebih dahulu
ALTER TABLE account_types 
ADD COLUMN IF NOT EXISTS user_id BIGINT;

-- 2. Jika ada data lama, isi user_id dengan ID default (misalnya ID user 1)
UPDATE account_types 
SET user_id = 1 
WHERE user_id IS NULL;

-- 3. Ubah kolom user_id menjadi NOT NULL
ALTER TABLE account_types 
ALTER COLUMN user_id SET NOT NULL;

-- 4. Hapus Foreign Key di tabel account_balance TERLEBIH DAHULU 
-- (agar constraint di account_types bebas dari ketergantungan)
ALTER TABLE account_balance 
DROP CONSTRAINT IF EXISTS account_balance_account_type_fkey;

-- 5. Baru hapus constraint UNIQUE lama di account_types
ALTER TABLE account_types 
DROP CONSTRAINT IF EXISTS account_types_type_code_key;

-- 6. Tambahkan constraint UNIQUE komposit baru pada account_types (user_id, type_code)
ALTER TABLE account_types 
ADD CONSTRAINT uq_user_type_code UNIQUE (user_id, type_code);

-- 7. Tambahkan Foreign Key ke tabel users pada account_types
ALTER TABLE account_types 
ADD CONSTRAINT fk_account_types_user 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- 8. Buat kembali Foreign Key komposit pada account_balance
ALTER TABLE account_balance 
ADD CONSTRAINT fk_user_account_type 
FOREIGN KEY (user_id, account_type) 
REFERENCES account_types(user_id, type_code) 
ON UPDATE CASCADE;