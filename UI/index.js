let cityCenter = document.getElementById("city-center")
let centerLat = document.getElementById("central-lat")
let centerLon = document.getElementById("central-lon")
let cLatVal = 0
let cLonVal = 0
let currentRadius = 5

const getCityCode = (cName) => {
    switch (cName) {
        case "mum":
            return 1
        case "ny":
            return 2
        default:
            break;
    }
}

const getFullCityName = (cityN)=>{
    switch (cityN) {
        case "mum":
            return "Mumbai"
        case "ny":
            return "New York"
        default:
            break;
    }
}

const fetchAndLoadData = async(cityName) =>{
    const response = await fetch(`http://localhost:8000/load-dummy?city=${getCityCode(cityName)}`)
    const data = await response.json()
    const dtt = document.getElementById("data-tag-table")

    console.log(data)
    cityCenter.innerText = getFullCityName(cityName)
    centerLat.innerText = data["lat"]
    cLatVal = parseFloat(data["lat"])
    centerLon.innerText = data["lon"]
    cLonVal = parseFloat(data["lon"])
    let placesData = data["placeData"]
    let allRowData = `<tr>
                            <th class="td-place">Place Name</th>
                            <th>Latitude</th>
                            <th>Longitude</th>
                        </tr>`
    placesData.map((place)=>{
        allRowData+=`<tr>
            <td class="td-place">${place.name}</td>
            <td>${place.latitude}</td>
            <td>${place.longitude}</td>
        </tr>`
    })
    dtt.innerHTML = allRowData

    fetchNearByLocations(cLatVal,cLonVal,currentRadius)
}

const handleRadiusSearch=()=>{
    const radInput = document.getElementById("rad-input")
    if(radInput.value !== currentRadius.toString()){
        currentRadius = parseFloat(radInput.value)
        fetchNearByLocations(cLatVal,cLonVal,currentRadius)
    }
}

const fetchNearByLocations = async(lat,lon,radius) =>{
    const fnRes = await fetch(`http://localhost:8000/find-nearby?lat=${lat}&lon=${lon}&radius=${radius}`)
    const fnData = await fnRes.json()
    console.log(fnData)
    const findNearbyResContainer = document.getElementById("find-nearby-results")
    let allRowData = `<div class="result-rows">
                    <p class="result-headers sn-col" >S.No.</p>
                    <p class="result-headers place-col" >Place</p>
                    <p class="result-headers lat-lan-col" >Latitude</p>
                    <p class="result-headers lat-lan-col" >Longitude</p>
                    <p class="result-headers dist-col" >Distance</p>
                    </div>`
    fnData.sort((a,b)=>a.distance - b.distance).map((fnRow,ind)=>{
        allRowData += `<div class="result-rows">
                    <p class="result-body sn-col">${ind+1}</p>
                    <p class="result-body place-col"><b>${fnRow.name}</b></p>
                    <p class="result-body lat-lan-col">${fnRow.latitude.toFixed(5)}</p>
                    <p class="result-body lat-lan-col">${fnRow.longitude.toFixed(5)}</p>
                    <p class="result-body dist-col">${fnRow.distance.toFixed(3)} KM</p>
                </div>`
    })
    findNearbyResContainer.innerHTML = allRowData
}


let allTags = ["mum","ny"]
const handleTagSwitch=(cName)=>{
        const tag = document.getElementById(cName)
        tag.classList.add('active-city')
        allTags.forEach(tagName => {
            if(tagName!==cName){
                const tagElement = document.getElementById(tagName)
                tagElement.classList.remove('active-city')
            }
        })
        currentRadius = 5
        const radInput = document.getElementById("rad-input")
        radInput.value = currentRadius
        fetchAndLoadData(cName)

}

fetchAndLoadData("mum")