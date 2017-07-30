<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Controller;


use Statusengine\Loader\HostLoaderInterface;
use Statusengine\ValueObjects\HostQueryOptions;
use Statusengine\ValueObjects\HostSearchQueryOptions;

class Host extends Controller {

    /**
     * @var HostLoaderInterface
     */
    private $HostLoader;

    /***
     * Host constructor.
     * @param HostLoaderInterface $HostLoader
     */
    public function __construct(HostLoaderInterface $HostLoader) {
        $this->HostLoader = $HostLoader;
    }

    /**
     * @return array
     */
    public function index(HostQueryOptions $queryOptions) {
        return $this->HostLoader->getHostList($queryOptions);
    }

    /**
     * @param HostQueryOptions $queryOptions
     * @return array
     */
    public function hostdetails(HostQueryOptions $queryOptions){
        return $this->HostLoader->getHostDetails($queryOptions);
    }

    /**
     * @param HostSearchQueryOptions $HostSearchQueryOptions
     * @return array
     */
    public function search(HostSearchQueryOptions $HostSearchQueryOptions){
        return $this->HostLoader->search($HostSearchQueryOptions);
    }


}