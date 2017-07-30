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

use Statusengine\Loader\Crate\ExternalCommandSaver;
use Statusengine\Saver\ExternalCommandSaverInterface;
use Statusengine\ValueObjects\ExternalCommandArgsQueryOptions;
use Statusengine\ValueObjects\ExternalCommandQueryOptions;

class ExternalCommand extends Controller {

    /**
     * @var ExternalCommandSaver
     */
    private $ExternalCommandSaver;

    /**
     * ExternalCommand constructor.
     * @param ExternalCommandSaverInterface $ExternalCommandSaver
     */
    public function __construct(ExternalCommandSaverInterface $ExternalCommandSaver) {
        $this->ExternalCommandSaver = $ExternalCommandSaver;
    }

    /**
     * @param ExternalCommandQueryOptions $ExternalCommandQueryOptions
     * @return bool
     */
    public function index(ExternalCommandQueryOptions $ExternalCommandQueryOptions) {
        $ExternalCommand = new \Statusengine\Generators\ExternalCommand($ExternalCommandQueryOptions);
        $commandAsStringOrArray = $ExternalCommand->getCommand();
        if(is_array($commandAsStringOrArray)){
            foreach($commandAsStringOrArray as $command){
                $this->ExternalCommandSaver->saveCommand($command, $ExternalCommandQueryOptions->getNodeName());
            }
            return true;
        }

        return $this->ExternalCommandSaver->saveCommand($commandAsStringOrArray, $ExternalCommandQueryOptions->getNodeName());

    }

    /**
     * @param ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions
     * @return bool
     */
    public function args(ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions) {
        $ExternalCommand = new \Statusengine\Generators\ExternalCommandArgs($ExternalCommandArgsQueryOptions);
        $commandAsStringOrArray = $ExternalCommand->getCommand();
        if(is_array($commandAsStringOrArray)){
            foreach($commandAsStringOrArray as $command){
                $this->ExternalCommandSaver->saveCommand($command, $ExternalCommandArgsQueryOptions->getNodeName());
            }
            return true;
        }

        return $this->ExternalCommandSaver->saveCommand($commandAsStringOrArray, $ExternalCommandArgsQueryOptions->getNodeName());

    }

}
