from typing import Optional

from fastapi import Header, HTTPException, Depends

from app.core.security import verify_token
from app.db.session import Session
from app.db.models import Ride


async def get_user_id(authorization: str = Header(None)) -> str:
    if authorization is None:
        raise HTTPException(status_code=401, detail="Authorization header missing")

    try:
        token = authorization.split("Bearer ")[1]
        payload = await verify_token(token)
        user_id = payload.get("id")
        return user_id
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid or expired token")


def get_current_user_id(user_id: str = Depends(get_user_id)):
    return user_id


def get_ride_id_by_user(db: Session, user_id: int):
    ride = db.query(Ride).filter(
        (Ride.passenger_id == user_id) | (Ride.driver_id == user_id)
        ).first()

    if ride:
        return ride.id
    return
