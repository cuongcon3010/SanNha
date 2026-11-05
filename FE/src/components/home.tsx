"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import {
  ArrowRight,
  MapPin,
  Search,
  CalendarDays,
  Clock3,
  CircleDot,
  MoveUpRight,
  Volleyball,
  Trophy,
  ShieldCheck,
  Zap,
  Check,
} from "lucide-react";
import { cities, filterVenues, sports, today, venues } from "@/lib/data";
import { Header } from "./header";
import { Arrow, Empty, Footer } from "./ui";
import { VenueCard } from "./venue-card";

function Pitch() {
  return (
    <div className="pitch-scene" aria-hidden="true">
      <div className="pitch-caption">
        <span className="status-dot" /> SÂN NHÀ / PLAY TOGETHER
      </div>
      <div className="pitch-frame">
        <div className="pitch">
          <div className="pitch-midline" />
          <div className="pitch-circle" />
          <div className="penalty penalty-left" />
          <div className="penalty penalty-right" />
          <div className="goal goal-left" />
          <div className="goal goal-right" />
          <span className="player p1" />
          <span className="player p2" />
          <span className="player p3" />
          <span className="player p4" />
          <span className="player p5" />
          <span className="player ball" />
        </div>
      </div>
      <div className="pitch-live">
        <span className="pulse-dot" />
        <div>
          <strong>Sẵn sàng cho trận đấu tiếp theo?</strong>
          <span>Kèo hay bắt đầu từ một sân tốt.</span>
        </div>
        <MoveUpRight size={22} />
      </div>
      <div className="pitch-index">
        21°01′ N &nbsp; 105°48′ E <span>HÀ NỘI, VIỆT NAM</span>
      </div>
    </div>
  );
}

export function Home() {
  const [sport, setSport] = useState("Tất cả");
  const [city, setCity] = useState("Hà Nội");
  const [query, setQuery] = useState("");
  const [date, setDate] = useState("");
  const [period, setPeriod] = useState("evening");
  const [searched, setSearched] = useState(false);
  const filtered = filterVenues(venues, sport, city, query);
  function search(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSearched(true);
    document
      .getElementById("san-noi-bat")
      ?.scrollIntoView({ behavior: "smooth" });
  }
  return (
    <>
      <Header />
      <main id="main-content">
        <section className="hero dotted-dark">
          <div className="container">
            <div className="hero-top">
              <div className="hero-copy">
                <span className="hero-eyebrow">
                  <span className="status-dot" />
                  RA SÂN. KẾT NỐI. HẾT MÌNH.
                </span>
                <h1>
                  Kèo đã có.
                  <br />
                  <span>Sân để chúng tôi.</span>
                </h1>
                <p>
                  Tìm sân đúng gu, chọn giờ đúng ý.
                  <br />
                  Đặt sân thể thao dễ dàng — sẵn sàng cho mọi cuộc chơi.
                </p>
                <div className="hero-steps">
                  <span>01 / TÌM SÂN</span>
                  <i />
                  <span>02 / CHỌN GIỜ</span>
                  <i />
                  <span>03 / RA SÂN</span>
                </div>
              </div>
              <Pitch />
            </div>
            <div className="sport-picker">
              <span className="eyebrow">CHỌN MÔN CỦA BẠN</span>
              <div className="sport-tabs">
                {sports.map((s, i) => (
                  <button
                    key={s}
                    className={sport === s ? "sport-tab active" : "sport-tab"}
                    onClick={() => setSport(s)}
                    aria-pressed={sport === s}
                  >
                    {i === 0 ? (
                      <Trophy size={17} />
                    ) : i === 1 ? (
                      <Volleyball size={17} />
                    ) : (
                      <CircleDot size={17} />
                    )}
                    {s}
                    {s === "Pickleball" && <span className="hot-dot" />}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </section>
        <div className="container search-wrap">
          <form className="search-panel" onSubmit={search}>
            <label>
              <span>
                <MapPin size={13} /> KHU VỰC
              </span>
              <select value={city} onChange={(e) => setCity(e.target.value)}>
                <option value="">Toàn quốc</option>
                {cities.map((c) => (
                  <option key={c}>{c}</option>
                ))}
              </select>
            </label>
            <label className="search-keyword">
              <span>
                <Search size={13} /> TÊN SÂN / ĐỊA ĐIỂM
              </span>
              <input
                placeholder="Bạn muốn chơi ở đâu?"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
            </label>
            <label>
              <span>
                <CalendarDays size={13} /> NGÀY CHƠI
              </span>
              <input
                type="date"
                aria-label="Ngày chơi"
                min={date ? today() : undefined}
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </label>
            <label>
              <span>
                <Clock3 size={13} /> KHUNG GIỜ
              </span>
              <select
                value={period}
                onChange={(e) => setPeriod(e.target.value)}
              >
                <option value="evening">17:00 – 22:00</option>
                <option value="all">Tất cả giờ demo</option>
              </select>
            </label>
            <button className="button search-submit" type="submit">
              Tìm sân <Arrow />
            </button>
          </form>
          <p className="quick-search">
            Gợi ý:{" "}
            {[
              ["Bóng đá", "Cầu Giấy"],
              ["Pickleball", "Mễ Trì"],
              ["Cầu lông", "Hà Đông"],
            ].map(([s, q]) => (
              <button
                key={s}
                onClick={() => {
                  setSport(s);
                  setQuery(q);
                  setCity("Hà Nội");
                  setSearched(true);
                  document
                    .getElementById("san-noi-bat")
                    ?.scrollIntoView({ behavior: "smooth" });
                }}
              >
                {s} {q}
                <Arrow />
              </button>
            ))}
          </p>
        </div>
        <section className="section container categories">
          <div className="section-heading">
            <div>
              <span className="eyebrow">KHÔNG CHỈ LÀ MỘT TRẬN ĐẤU</span>
              <h2>Một cú chạm. Mọi sân chơi.</h2>
              <p>
                Chọn bộ môn bạn yêu thích, tìm đồng đội và bắt đầu cuộc chơi.
              </p>
            </div>
            <span className="section-index">01 — KHÁM PHÁ</span>
          </div>
          <div className="category-grid">
            {[
              ["Bóng đá", "Sân 5, 7 & 11 người"],
              ["Cầu lông", "Sàn chuẩn, chơi cực đã"],
              ["Pickleball", "Bắt nhịp cuộc chơi mới"],
              ["Tennis", "Bứt phá từng đường bóng"],
            ].map(([s, caption], i) => (
              <button
                className={`category-card ${sport === s || (sport === "Tất cả" && i === 0) ? "active" : ""}`}
                key={s}
                onClick={() => {
                  setSport(s);
                  setQuery("");
                  document
                    .getElementById("san-noi-bat")
                    ?.scrollIntoView({ behavior: "smooth" });
                }}
              >
                <span className="category-number">
                  0{i + 1}
                  <CircleDot size={33} strokeWidth={1.2} />
                </span>
                <span className="category-name">
                  {s}
                  <Arrow />
                </span>
                <span className="category-caption">{caption}</span>
              </button>
            ))}
          </div>
        </section>
        <section className="section container featured" id="san-noi-bat">
          <div className="section-heading">
            <div>
              <span className="eyebrow">SÂN ĐẸP. GIÁ RÕ. SẴN SÀNG CHƠI.</span>
              <h2>Điểm hẹn cho kèo hôm nay</h2>
            </div>
            <button
              className="text-link"
              onClick={() => {
                setSport("Tất cả");
                setCity("");
                setQuery("");
              }}
            >
              Xem tất cả <Arrow />
            </button>
          </div>
          <div className="filter-row">
            {sports.slice(0, 5).map((s) => (
              <button
                key={s}
                aria-pressed={sport === s}
                className={sport === s ? "filter-chip active" : "filter-chip"}
                onClick={() => setSport(s)}
              >
                {s === "Tất cả" ? "Tất cả sân" : s}
              </button>
            ))}
            <span className="result-count">{filtered.length} sân phù hợp</span>
          </div>
          {searched && (
            <p className="search-feedback" role="status">
              Kết quả trong bộ dữ liệu demo
              {date ? ` cho ngày ${date.split("-").reverse().join("/")}` : ""}.
              Chọn sân để xem lịch{" "}
              {period === "evening" ? "17:00 – 22:00" : "16:00 – 22:00"}.
            </p>
          )}
          <div className="venue-grid">
            {filtered.map((v) => (
              <VenueCard key={v.id} venue={v} date={date} period={period} />
            ))}
          </div>
          {filtered.length === 0 && (
            <Empty
              title="Chưa tìm thấy sân phù hợp"
              description="Dữ liệu mẫu hiện có 3 sân tại Hà Nội. Thử chọn môn hoặc khu vực khác."
            />
          )}
          <p className="photo-note">
            Ảnh thể thao minh họa · Giá và đánh giá là dữ liệu mẫu.
          </p>
        </section>
        <section className="how-section dotted-dark" id="cach-dat-san">
          <div className="container">
            <div className="section-heading">
              <div>
                <span className="eyebrow">ÍT THAO TÁC. NHIỀU CUỘC VUI.</span>
                <h2>Từ màn hình đến sân chơi.</h2>
              </div>
              <span className="section-index">03 BƯỚC. MỘT CUỘC CHƠI.</span>
            </div>
            <div className="how-grid">
              {[
                [
                  "Tìm sân hợp gu",
                  "Lọc theo bộ môn, khu vực và khung giờ.",
                  Search,
                ],
                [
                  "Chốt giờ, giữ sân",
                  "Xem giá rõ ràng, chọn lịch phù hợp.",
                  Zap,
                ],
                [
                  "Xách giày ra sân",
                  "Nhận xác nhận và sẵn sàng hết mình.",
                  Check,
                ],
              ].map(([title, caption, Icon], i) => {
                const StepIcon = Icon as typeof Search;
                return (
                  <div className="how-card" key={i}>
                    <span className="how-number">0{i + 1}</span>
                    <div>
                      <StepIcon size={20} />
                      <h3>{title as string}</h3>
                      <p>{caption as string}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </section>
        <section className="owner-cta container">
          <div>
            <span className="eyebrow">DÀNH CHO CHỦ SÂN</span>
            <h2>
              Bạn có sân.
              <br className="mobile-only" /> Chúng tôi có người chơi.
            </h2>
            <p>Đưa sân của bạn lên Sân Nhà và quản lý lịch dễ dàng hơn.</p>
          </div>
          <Link className="button" href="/dang-ky?role=owner">
            Kết nối chủ sân <ArrowRight size={17} />
          </Link>
        </section>
        <div className="container home-demo">
          <ShieldCheck size={15} />
          Phiên bản giao diện demo — chưa kết nối lịch sân và thanh toán thực
          tế.
        </div>
      </main>
      <Footer />
    </>
  );
}
