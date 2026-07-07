from pydantic import BaseModel



class IncidentCreate(BaseModel):

    title:str

    description:str



class IncidentResponse(BaseModel):

    id:int

    title:str

    description:str

    category:str | None

    priority:str

    status:str


    class Config:

        from_attributes=True