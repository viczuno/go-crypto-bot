import { ImageResponse } from "next/og";

export const alt =
  "CRYPTO Market Observatory — independent cryptocurrency market research";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpenGraphImage() {
  return new ImageResponse(
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "68px 76px",
        backgroundColor: "#0b1019",
        color: "#e7edf5",
        fontFamily: "Arial, sans-serif",
      }}
    >
      <div style={{ display: "flex", flexDirection: "column", width: 690 }}>
        <div
          style={{ display: "flex", alignItems: "center", marginBottom: 54 }}
        >
          <svg width="44" height="44" viewBox="0 0 48 48" fill="none">
            <path
              d="M37 10H23a14 14 0 1 0 0 28h14"
              stroke="#39d5b4"
              strokeWidth="7"
              strokeLinecap="round"
            />
            <path
              d="M14 25h6l3-6 4 11 3-5h4"
              stroke="#39d5b4"
              strokeWidth="3.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              marginLeft: 13,
            }}
          >
            <span style={{ fontSize: 19, fontWeight: 700, letterSpacing: 2.5 }}>
              CRYPTO
            </span>
            <span
              style={{
                color: "#91a0b6",
                fontSize: 8,
                fontWeight: 600,
                letterSpacing: 1.6,
                marginTop: 4,
              }}
            >
              MARKET OBSERVATORY
            </span>
          </div>
        </div>
        <div
          style={{
            color: "#39d5b4",
            fontSize: 13,
            fontWeight: 600,
            letterSpacing: 2.4,
            marginBottom: 18,
          }}
        >
          INDEPENDENT MARKET RESEARCH
        </div>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            fontSize: 58,
            fontWeight: 600,
            letterSpacing: -2.2,
            lineHeight: 1.1,
          }}
        >
          <span>Read the market</span>
          <span>with more clarity.</span>
        </div>
        <div
          style={{
            color: "#91a0b6",
            fontSize: 17,
            marginTop: 25,
          }}
        >
          Sampled prices · Go-powered technical analysis
        </div>
      </div>
      <div
        style={{
          width: 300,
          height: 330,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          border: "1px solid #233044",
          borderRadius: 24,
          backgroundColor: "#111925",
        }}
      >
        <svg width="176" height="176" viewBox="0 0 48 48" fill="none">
          <path
            d="M37 10H23a14 14 0 1 0 0 28h14"
            stroke="#39d5b4"
            strokeWidth="7"
            strokeLinecap="round"
          />
          <path
            d="M14 25h6l3-6 4 11 3-5h4"
            stroke="#39d5b4"
            strokeWidth="3.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        <div
          style={{
            color: "#91a0b6",
            fontSize: 9,
            fontWeight: 600,
            letterSpacing: 2,
            marginTop: 26,
          }}
        >
          OBSERVE · ANALYZE · COMPARE
        </div>
      </div>
    </div>,
    size,
  );
}
