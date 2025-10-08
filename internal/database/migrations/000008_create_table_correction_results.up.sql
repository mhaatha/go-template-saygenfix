CREATE TABLE correction_results (
    student_answer_id UUID NOT NULL,
    question TEXT NOT NULL,
    student_answer TEXT NOT NULL,
    score INTEGER NOT NULL,
    feedback VARCHAR(255) NOT NULL,
    
    CONSTRAINT fk_student_answer_id
        FOREIGN KEY (student_answer_id)
        REFERENCES student_answers(id)
        ON DELETE CASCADE
);
