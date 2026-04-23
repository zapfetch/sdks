#!/usr/bin/env python3
"""
Minimal examples for Zapfetch v2.
"""

import os
from dotenv import load_dotenv
from zapfetch import Zapfetch
 

load_dotenv()

def main():
    api_key = os.getenv("ZAPFETCH_API_KEY")
    if not api_key:
        raise ValueError("ZAPFETCH_API_KEY is not set")
    
    api_url = os.getenv("ZAPFETCH_API_URL")
    if not api_url:
        raise ValueError("ZAPFETCH_API_URL is not set")

    zapfetch = Zapfetch(api_key=api_key, api_url=api_url)

    # Scrape
    doc = zapfetch.scrape("https://docs.zapfetch.com", formats=["markdown"])
    print("scrape:", doc.markdown)
    # doc.metadata_dict is a dict, doc.metadata_typed is a DocumentMetadata object
    print(doc.metadata_dict.get("source_url"))
    print('metadata_dict.get("title"):', doc.metadata_dict.get("title"))
    print("metadata_typed.title:", doc.metadata_typed.title)
    print("metadata.title", doc.metadata.title if doc.metadata else None)


    # Crawl (waits until terminal state)
    crawl_job = zapfetch.crawl("https://docs.zapfetch.com", limit=3, poll_interval=1, timeout=120)
    print("crawl:", crawl_job.status, crawl_job.completed, "/", crawl_job.total)

    # Batch scrape
    batch = zapfetch.batch_scrape([
        "https://docs.zapfetch.com",
        "https://zapfetch.com",
    ], formats=["markdown"], poll_interval=1, wait_timeout=120)
    print("batch:", batch.status, batch.completed, "/", batch.total)

    # Search
    search_response = zapfetch.search(query="What is the capital of France?", limit=5)
    print("search web results:", len(getattr(search_response, "web", []) or []))

    # Map
    map_response = zapfetch.map("https://zapfetch.com")
    print("map links:", len(getattr(map_response, "links", []) or []))

if __name__ == "__main__":
    main()
