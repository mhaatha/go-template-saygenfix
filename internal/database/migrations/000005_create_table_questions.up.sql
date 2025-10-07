CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    exam_id VARCHAR(100) NOT NULL,
    
    CONSTRAINT fk_exam
        FOREIGN KEY(exam_id) 
        REFERENCES exams(id)
        ON DELETE CASCADE
);