Original_urls:
- id: int (primary key)
- url: string
- created_at: timestamp
- updated_at: timestamp

Short_url table:
- id: int (primary key)
- code: string
- original_url_id: int
- created_at: timestamp
- updated_at: timestamp