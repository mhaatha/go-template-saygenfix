CREATE TABLE student_answers (
    -- ID unik untuk setiap jawaban.
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- ID sesi pengerjaan ujian, merujuk ke tabel 'exam_attempts'.
    exam_attempt_id UUID NOT NULL,

    -- ID pertanyaan yang dijawab, merujuk ke tabel 'questions'.
    question_id UUID NOT NULL,

    -- Jawaban mahasiswa dalam format teks.
    student_answer TEXT NOT NULL DEFAULT '',
    
    score INTEGER NOT NULL DEFAULT 0,
    feedback VARCHAR(255) NOT NULL DEFAULT '',
    question_max_score INTEGER NOT NULL DEFAULT 0,
    

    -- Mendefinisikan foreign key.
    FOREIGN KEY (exam_attempt_id) REFERENCES exam_attempts(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
);