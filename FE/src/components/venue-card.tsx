"use client";

import Image from "next/image";
import Link from "next/link";
import { Heart, MapPin, Star } from "lucide-react";
import { money, type Venue } from "@/lib/data";
import { useDemo } from "@/lib/demo-store";
import { Arrow } from "./ui";

export function VenueCard({
  venue,
  date = "",
  period = "evening",
}: {
  venue: Venue;
  date?: string;
  period?: string;
}) {
  const { favorites, toggleFavorite } = useDemo();
  const liked = favorites.includes(venue.id);
  const href = `/san/${venue.id}?${new URLSearchParams({ ...(date ? { date } : {}), period })}`;
  return (
    <article className="venue-card">
      <div className="venue-photo">
        <Link href={href} aria-label={`Xem ${venue.name}`}>
          <Image
            src={venue.image}
            alt={`Ảnh thể thao minh họa cho ${venue.name}`}
            fill
            sizes="(max-width: 650px) 100vw, (max-width: 950px) 50vw, 33vw"
          />
        </Link>
        <span className="availability">
          <span className="status-dot" />
          Lịch sân demo
        </span>
        <button
          className={`favorite-button ${liked ? "liked" : ""}`}
          aria-label={`${liked ? "Bỏ thích" : "Yêu thích"} ${venue.name}`}
          aria-pressed={liked}
          onClick={() => toggleFavorite(venue.id)}
        >
          <Heart size={17} fill={liked ? "currentColor" : "none"} />
        </button>
        <span className="photo-sport">{venue.sport}</span>
      </div>
      <div className="venue-body">
        <div className="rating">
          <Star size={13} fill="currentColor" /> <strong>{venue.rating}</strong>
          <span>· 128 đánh giá mẫu</span>
        </div>
        <h3>
          <Link href={href}>{venue.name}</Link>
        </h3>
        <p className="venue-address">
          <MapPin size={14} />
          {venue.address}
        </p>
        <div className="venue-bottom">
          <div>
            <small>Giá từ</small>
            <p>
              <strong>{money(venue.price)}</strong>
              <span> / giờ</span>
            </p>
          </div>
          <Link href={href} className="button">
            Xem lịch <Arrow />
          </Link>
        </div>
      </div>
    </article>
  );
}
