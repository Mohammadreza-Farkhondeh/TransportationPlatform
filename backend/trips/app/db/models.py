from sqlalchemy import Column, String, Integer, Float, DateTime, func
from sqlalchemy import Enum as PgEnum
from enum import Enum

from app.db.base import Base


class RideStatus(Enum):
    NOT_ACCEPTED = 0
    ACCEPTED_NOT_PICKED = 1
    IN_PROGRESS = 2
    DONE = 3
    NOT_ACCEPTED_IN_TIME = -1
    CANCEL_BY_PASSENGER = -2
    CANCEL_BY_DRIVER = -3


class Ride(Base):

    id = Column(Integer, primary_key=True, index=True)
    passenger_id = Column(String, index=True)
    driver_id = Column(String, index=True)
    origin_lat = Column(Float)
    origin_lon = Column(Float)
    destination_lat = Column(Float)
    destination_lon = Column(Float)
    status = Column(PgEnum(RideStatus), nullable=False)
    date_requested = Column(DateTime, default=func.now())
    date_accepted = Column(DateTime, default=None)
    date_arrived = Column(DateTime, default=None)
