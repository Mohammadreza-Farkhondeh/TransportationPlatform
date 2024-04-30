from typing import Optional
from pydantic import BaseModel
from datetime import datetime


class RideBase(BaseModel):
    origin_lat: float
    origin_lon: float
    destination_lat: float
    destination_lon: float


class RideCreate(RideBase):
    pass


class RideUpdate(RideBase):
    status: int


class Ride(RideBase):
    id: int
    passenger_id: str
    driver_id: Optional[str] = None
    status: int
    date_requested: datetime
    date_accepted: Optional[datetime]
    date_arrived: Optional[datetime]
