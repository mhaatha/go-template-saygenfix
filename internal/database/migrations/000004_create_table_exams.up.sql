CREATE TABLE exams (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    year SMALLINT,
    teacher_id UUID NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign Key ke tabel users (untuk guru yang membuat)
    -- Jika guru (user) dihapus, ujian yang dibuatnya juga akan terhapus
    CONSTRAINT fk_teacher
        FOREIGN KEY(teacher_id) 
        REFERENCES users(id)
        ON DELETE CASCADE
);