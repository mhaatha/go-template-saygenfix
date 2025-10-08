CREATE TABLE exam_attempts (
    -- ID unik untuk setiap percobaan ujian, menggunakan tipe UUID.
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- ID mahasiswa yang mengerjakan ujian, merujuk ke tabel 'users'.
    student_id UUID NOT NULL,

    -- ID ujian yang dikerjakan, merujuk ke tabel 'exams'.
    -- Menggunakan VARCHAR(100) agar sesuai dengan tipe data di tabel 'exams'.
    exam_id VARCHAR(100) NOT NULL,

    -- Nilai akhir yang didapat mahasiswa untuk ujian ini.
    score SMALLINT NOT NULL DEFAULT 0,

    -- Waktu kapan mahasiswa mulai mengerjakan ujian.
    started_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Waktu kapan mahasiswa menyelesaikan ujian.
    completed_at TIMESTAMP,

    -- Mendefinisikan foreign key untuk memastikan integritas data.
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (exam_id) REFERENCES exams(id) ON DELETE CASCADE
);